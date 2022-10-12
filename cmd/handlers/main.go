package handlers

import (
	"fmt"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
)

var validate *validator.Validate
var enTrans ut.Translator

func init(){
	validate = validator.New()
	english := en.New()
	uni := ut.New(english, english)
	enTrans, _ = uni.GetTranslator("en")
	_ = enTranslations.RegisterDefaultTranslations(validate, enTrans)

}

func translateError(err error, trans ut.Translator) (errs []error) {
	if err == nil {
	  return nil
	}
	validatorErrs := err.(validator.ValidationErrors)
	for _, e := range validatorErrs {
	  translatedErr := fmt.Errorf(e.Translate(trans))
	  errs = append(errs, translatedErr)
	}
	return errs
}

func stringfyJSONErrArr(errs []error) []string {
	strErrors := make([]string, len(errs))

	for i, err := range errs {
		strErrors[i] = err.Error()
	}

	return strErrors
}