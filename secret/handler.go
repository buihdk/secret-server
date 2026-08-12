package secret

import (
	"net/http"
	"strconv"
	"time"

	"secretserver/internal/crypto"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/labstack/echo"
)

func GetSecret(c echo.Context) error {
	hash := c.Param("hash")

	db, err := collection()
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	// Atomically claim one view. The filter requires remainingViews > 0 so two
	// concurrent reads cannot both succeed on a secret with one view left.
	// $inc of -1 decrements as part of the same operation.
	filter := bson.D{
		{Key: "hash", Value: hash},
		{Key: "remainingViews", Value: bson.D{{Key: "$gt", Value: 0}}},
	}
	update := bson.D{{Key: "$inc", Value: bson.D{{Key: "remainingViews", Value: -1}}}}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.Before)

	result := new(Secret)
	err = db.FindOneAndUpdate(c.Request().Context(), filter, update, opts).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return c.String(http.StatusNotFound, "Secret not found")
		}
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	// Check time expiry (zero ExpiresAt means never expires).
	if !result.ExpiresAt.IsZero() && time.Now().After(result.ExpiresAt) {
		db.DeleteOne(c.Request().Context(), bson.D{{Key: "hash", Value: hash}}) //nolint:errcheck
		return c.String(http.StatusNotFound, "Secret not found")
	}

	// result holds the pre-update state. If it had one view left, this read
	// consumed the last one, so delete the document.
	if result.RemainingViews <= 1 {
		if _, err := db.DeleteOne(c.Request().Context(), bson.D{{Key: "hash", Value: hash}}); err != nil {
			c.Logger().Error(err)
			return c.String(http.StatusInternalServerError, "Internal server error")
		}
	}

	// Report the count after consuming this view.
	result.RemainingViews--

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

	// Parse directly as int32 so a value that would overflow the stored field
	// is rejected instead of wrapping to a negative or truncated count.
	expireAfterViews, err := strconv.ParseInt(expireAfterViewsR, 10, 32)
	if err != nil || expireAfterViews <= 0 {
		c.Logger().Info("Invalid expireAfterViews")
		return c.String(http.StatusMethodNotAllowed, "Invalid input")
	}

	minutes, err := strconv.ParseInt(expireAfterR, 10, 64)
	if err != nil || minutes < 0 {
		c.Logger().Info("Invalid expireAfter")
		return c.String(http.StatusMethodNotAllowed, "Invalid input")
	}

	// 0 = never expires; otherwise offset from now in minutes.
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

	db, err := collection()
	if err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "Internal server error")
	}

	if _, err = db.InsertOne(c.Request().Context(), secret); err != nil {
		c.Logger().Error(err)
		return c.String(http.StatusInternalServerError, "Failed to store secret")
	}

	// Return the stored record directly. Do not read it back through
	// GetSecret, which would consume a view and delete a one-view secret
	// before the recipient can fetch it. Return the plaintext the caller sent.
	secret.SecretText = secretText

	return c.JSON(http.StatusOK, secret)
}
