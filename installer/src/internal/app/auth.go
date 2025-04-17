package app

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func (app *App) createJwtToken(userId string) (string, error) {
	jwtKeyBuffer, err := app.JwtKeyEnclave.Open()
	if err != nil {
		return "", err
	}
	defer jwtKeyBuffer.Destroy()

	now := time.Now()
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss": jwtIssuer,
			"exp": now.Add(jwtTokenLifetime).Unix(),
			"iat": now.Unix(),
			"sub": userId,
		},
	)
	return jwtToken.SignedString(jwtKeyBuffer.Bytes())
}

func (app *App) verifyJwtToken(tokenString string) (*string, error) {
	jwtKeyBuffer, err := app.JwtKeyEnclave.Open()
	if err != nil {
		return nil, err
	}
	defer jwtKeyBuffer.Destroy()

	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {
			return jwtKeyBuffer.Bytes(), nil
		},
		jwt.WithIssuer(jwtIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid jwt token")
	}

	userId, err := token.Claims.GetSubject()
	if err != nil {
		return nil, err
	}

	return &userId, nil
}

func (app *App) verifyInstallerToken(token string) error {
	tokenBuffer, err := app.AiidaInstaller.Encryptor.InstallerTokenEnclave.Open()
	if err != nil {
		return err
	}
	defer tokenBuffer.Destroy()

	if tokenBuffer.String() != token {
		return errors.New("invalid installer token")
	}

	return nil
}
