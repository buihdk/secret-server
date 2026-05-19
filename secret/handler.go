package secret

import (
	"net/http"
	"strconv"
	"time"

	"secretserver/internal/crypto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/labstack/echo"
)

func GetSecret(c echo.Context) error {
	hash := c.Param("hash")

	result := new(Secret)
	err := db.FindOne(c.Request().Context(), bson.D{{Key: "hash", Value: hash}}).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.String(http.StatusNotFound, "Secret not found")
		}
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	// Check time expiry (zero ExpiresAt means never expires)
	if !result.ExpiresAt.IsZero() && time.Now().After(result.ExpiresAt) {
		db.DeleteOne(c.Request().Context(), bson.D{{Key: "hash", Value: hash}}) //nolint:errcheck
		return c.String(http.StatusNotFound, "Secret not found")
	}

	// Check view count
	if result.RemainingViews <= 0 {
		db.DeleteOne(c.Request().Context(), bson.D{{Key: "hash", Value: hash}}) //nolint:errcheck
		return c.String(http.StatusNotFound, "Secret not found")
	}

	// Consume one view
	result.RemainingViews--
	if result.RemainingViews == 0 {
		if _, err := db.DeleteOne(c.Request().Context(), bson.D{{Key: "hash", Value: hash}}); err != nil {
			c.Logger().Error(err)
			return c.String(http.StatusInternalServerError, "Internal server error")
		}
	} else {
		update := bson.D{{Key: "$set", Value: bson.D{{Key: "remainingViews", Value: result.RemainingViews}}}}
		if _, err := db.UpdateOne(c.Request().Context(), bson.D{{Key: "hash", Value: hash}}, update); err != nil {
			c.Logger().Error(err)
			return c.String(http.StatusInternalServerError, "Internal server error")
		}
	}

	plaintext, err := crypto.Decrypt(result.SecretText)
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "Failed to decrypt secret")
	}
	result.SecretText = plaintext

	return c.JSON(http.StatusOK, result)
}

func AddSecret(c echo.Context) error {
	secretText := c.FormValue("secret")
	expireAfterViewsR := c.FormValue("expireAfterViews")
	expireAfterR := c.FormValue("expireAfter")

	expireAfterViews, err := strconv.Atoi(expireAfterViewsR)
	if err != nil || expireAfterViews <= 0 {
		c.Logger().Info("Invalid expireAfterViews")
		return c.String(http.StatusMethodNotAllowed, "Invalid input")
	}

	minutes, err := strconv.ParseInt(expireAfterR, 10, 64)
	if err != nil {
		c.Logger().Info("Invalid expireAfter")
		return c.String(http.StatusMethodNotAllowed, "Invalid input")
	}

	// 0 = never expires; otherwise offset from now in minutes
	var expiresAt time.Time
	if minutes > 0 {
		expiresAt = time.Now().Add(time.Duration(minutes) * time.Minute)
	}

	encryptedText, err := crypto.Encrypt(secretText)
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "Failed to encrypt secret")
	}

	secret := Secret{
		SecretText:     encryptedText,
		RemainingViews: int32(expireAfterViews),
		ExpiresAt:      expiresAt,
		CreatedAt:      time.Now(),
	}
	secret.DoHash()

	if _, err = db.InsertOne(c.Request().Context(), secret); err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "Failed to store secret")
	}

	return GetSecret(addCtxValue(c, "hash", secret.Hash))
}

func addCtxValue(c echo.Context, key, value string) echo.Context {
	k := c.ParamNames()
	k = append(k, key)

	v := c.ParamValues()
	v = append(v, value)

	c.SetParamNames(k...)
	c.SetParamValues(v...)

	return c
}
