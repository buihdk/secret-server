package secret

import (
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/labstack/echo"
)

func GetSecret(c echo.Context) error {
	hash := c.Param("hash")
	findQuery := bson.D{{"hash", hash}}

	result := new(Secret)
	err := db.FindOne(c.Request().Context(), findQuery).Decode(result)
	if err != nil {
		c.Logger().Error(err)
		return c.String(400, "Secret not found")
	}

	if err := c.Bind(result); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, result)
}

func AddSecret(c echo.Context) error {
	secretText := c.FormValue("secret")
	expireAfterViewsR := c.FormValue("expireAfterViews")
	expireAfterR := c.FormValue("expireAfter")

	// parse
	expireAfterViews, err := strconv.Atoi(expireAfterViewsR)
	if err != nil {
		c.Logger().Info("Invalid expireAfterViews")
		return c.String(405, "Invalid input")
	}

	i, err := strconv.ParseInt(expireAfterR, 10, 64)
	if err != nil {
		c.Logger().Info("Invalid expireAfter")
		return c.String(405, "Invalid input")
	}
	// time must be in second
	expireAfter := time.Unix(i, 0)

	secret := Secret{
		SecretText:     secretText,
		RemainingViews: int32(expireAfterViews),
		ExpiresAt:      expireAfter,
		CreatedAt:      time.Now(),
	}
	secret.DoHash()

	// insert
	_, err = db.InsertOne(c.Request().Context(), secret)
	if err != nil {
		return c.String(405, "Invalid input")
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
