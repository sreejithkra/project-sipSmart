package helper

import (
	"fmt"
	"sip_Smart/database"
	"sip_Smart/models"

	"github.com/go-playground/validator"
)

func Validate(data interface{}) error {

	validate := validator.New()

	err := validate.Struct(data)
	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {

			switch e.Tag() {

			case "email":
				emailValue := e.Value().(string)
				var count int64
				database.Db.Model(&models.User{}).Where("email = ?", emailValue).Count(&count)
				if count > 0 {
					return fmt.Errorf("%s is already in use", e.Field())
				}
				return fmt.Errorf("%s is not a valid email address", e.Field())

			case "url":
				return fmt.Errorf("%s is not a valid url format", e.Field())

			case "max":
				return fmt.Errorf("%s exceeds the maximum length %s", e.Field(), e.Param())

			case "numeric":
				return fmt.Errorf("%s should contain only digits", e.Field())

			case "required":
				return fmt.Errorf("%s is required", e.Field())

			case "len":
				return fmt.Errorf("%s shouls have a length of %s", e.Field(), e.Param())

			case "gt":
				return fmt.Errorf("%s must be greater than zero", e.Field())

			case "min":
				return fmt.Errorf("%s shouls have a minimum length of %s", e.Field(), e.Param())

			default:
				return fmt.Errorf("validation error for field %s", e.Field())
			}

		}

	}
	return nil
}
