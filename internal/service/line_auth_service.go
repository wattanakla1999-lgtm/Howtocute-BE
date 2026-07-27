package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"nailly-back-end/internal/apperror"
	"nailly-back-end/internal/model"
	"net/http"
	"net/url"
	"strings"

	"gorm.io/gorm"
)

const (
	lineVerifyURL            = "https://api.line.me/oauth2/v2.1/verify"
	lineCustomerFallbackName = "LINE Customer"
)

type CustomerStore interface {
	FindByID(id string) (model.User, error)
	FindByLineUserID(lineUserID string) (model.User, error)
	Create(user *model.User) error
	Save(user *model.User) error
}

type LineTokenClaims struct {
	Subject  string
	Audience string
	Name     string
	Picture  string
}

type LineTokenVerifier interface {
	Verify(ctx context.Context, idToken, channelID string) (LineTokenClaims, error)
}

type HTTPLineTokenVerifier struct {
	client    *http.Client
	verifyURL string
}

func NewHTTPLineTokenVerifier(client *http.Client) *HTTPLineTokenVerifier {
	if client == nil {
		client = http.DefaultClient
	}
	return &HTTPLineTokenVerifier{client: client, verifyURL: lineVerifyURL}
}

func (v *HTTPLineTokenVerifier) Verify(ctx context.Context, idToken, channelID string) (LineTokenClaims, error) {
	form := url.Values{}
	form.Set("id_token", idToken)
	form.Set("client_id", channelID)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, v.verifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return LineTokenClaims{}, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := v.client.Do(request)
	if err != nil {
		log.Printf("LINE token verify request failed channel_id=%s error=%v", maskChannelID(channelID), err)
		return LineTokenClaims{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return LineTokenClaims{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		log.Printf(
			"LINE token verify failed status=%d channel_id=%s response=%s",
			response.StatusCode,
			maskChannelID(channelID),
			safeLogBody(body),
		)
		return LineTokenClaims{}, apperror.Unauthorized("invalid LINE token", apperror.ErrValidation)
	}

	var payload struct {
		Subject  string `json:"sub"`
		Audience string `json:"aud"`
		Name     string `json:"name"`
		Picture  string `json:"picture"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return LineTokenClaims{}, err
	}

	return LineTokenClaims{
		Subject:  payload.Subject,
		Audience: payload.Audience,
		Name:     payload.Name,
		Picture:  payload.Picture,
	}, nil
}

type LineAuthResult struct {
	Customer model.User
	Token    string
}

type LineAuthService struct {
	repo       CustomerStore
	verifier   LineTokenVerifier
	jwtManager *CustomerJWTManager
	channelID  string
}

func NewLineAuthService(repo CustomerStore, verifier LineTokenVerifier, jwtManager *CustomerJWTManager, channelID string) *LineAuthService {
	return &LineAuthService{
		repo:       repo,
		verifier:   verifier,
		jwtManager: jwtManager,
		channelID:  strings.TrimSpace(channelID),
	}
}

func (s *LineAuthService) Login(ctx context.Context, idToken string) (LineAuthResult, error) {
	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return LineAuthResult{}, apperror.BadRequest("idToken is required", apperror.ErrValidation)
	}
	if s.channelID == "" {
		return LineAuthResult{}, apperror.Internal("LINE_LOGIN_CHANNEL_ID is not configured", errors.New("missing LINE_LOGIN_CHANNEL_ID"))
	}
	if s.jwtManager == nil || len(s.jwtManager.secret) == 0 {
		return LineAuthResult{}, apperror.Internal("CUSTOMER_JWT_SECRET is not configured", errors.New("missing CUSTOMER_JWT_SECRET"))
	}
	if s.verifier == nil {
		return LineAuthResult{}, apperror.Internal("LINE token verifier is not configured", errors.New("missing LINE verifier"))
	}

	claims, err := s.verifier.Verify(ctx, idToken, s.channelID)
	if err != nil {
		var appErr *apperror.AppError
		if errors.As(err, &appErr) {
			return LineAuthResult{}, err
		}
		return LineAuthResult{}, apperror.Unauthorized("invalid LINE token", err)
	}
	if strings.TrimSpace(claims.Audience) != s.channelID {
		log.Printf(
			"LINE token audience mismatch expected_channel_id=%s token_audience=%s",
			maskChannelID(s.channelID),
			maskChannelID(claims.Audience),
		)
		return LineAuthResult{}, apperror.Unauthorized(fmt.Sprintf("invalid LINE token: Channel ID mismatch (backend expects %s, but token is for %s)", s.channelID, claims.Audience), apperror.ErrValidation)
	}
	lineUserID := strings.TrimSpace(claims.Subject)
	if lineUserID == "" {
		log.Printf("LINE token subject missing channel_id=%s", maskChannelID(s.channelID))
		return LineAuthResult{}, apperror.Unauthorized("invalid LINE token", apperror.ErrValidation)
	}

	name := strings.TrimSpace(claims.Name)
	if name == "" {
		name = lineCustomerFallbackName
	}

	customer, err := s.findOrCreateCustomer(lineUserID, name, strings.TrimSpace(claims.Picture))
	if err != nil {
		return LineAuthResult{}, err
	}

	token, _, err := s.jwtManager.Generate(customer)
	if err != nil {
		return LineAuthResult{}, apperror.Internal("could not create customer token", err)
	}

	return LineAuthResult{Customer: customer, Token: token}, nil
}

func (s *LineAuthService) findOrCreateCustomer(lineUserID, name, pictureURL string) (model.User, error) {
	customer, err := s.repo.FindByLineUserID(lineUserID)
	if err == nil {
		customer.Name = name
		customer.PictureURL = pictureURL
		if err := s.repo.Save(&customer); err != nil {
			return model.User{}, err
		}
		return customer, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.User{}, err
	}

	lineUserIDPtr := lineUserID
	customer = model.User{
		Name:       name,
		Email:      fmt.Sprintf("%s@line.nailly.local", strings.ToLower(lineUserID)),
		LineUserID: &lineUserIDPtr,
		PictureURL: pictureURL,
	}
	if err := s.repo.Create(&customer); err != nil {
		return model.User{}, err
	}
	return customer, nil
}

func maskChannelID(channelID string) string {
	channelID = strings.TrimSpace(channelID)
	if len(channelID) <= 4 {
		return "****"
	}
	return channelID[:4] + strings.Repeat("*", len(channelID)-4)
}

func safeLogBody(body []byte) string {
	const maxLogBody = 512
	text := strings.TrimSpace(string(body))
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	if len(text) > maxLogBody {
		return text[:maxLogBody] + "...[truncated]"
	}
	return text
}
