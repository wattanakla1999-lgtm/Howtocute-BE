package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
		return LineTokenClaims{}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return LineTokenClaims{}, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
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
		return LineAuthResult{}, apperror.Unauthorized("invalid LINE token", apperror.ErrValidation)
	}
	lineUserID := strings.TrimSpace(claims.Subject)
	if lineUserID == "" {
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
