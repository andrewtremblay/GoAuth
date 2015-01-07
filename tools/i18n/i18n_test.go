package i18n

import (
	"github.com/stretchr/testify/assert"
	"github.com/vaughan0/go-ini"
	"testing"
)

func getIniFile() ini.File {
	f, err := ini.LoadFile("./files/en-US.ini")
	if err != nil {
		panic(err.Error())
	}
	return f
}

func getIniGermanFile() ini.File {
	f, err := ini.LoadFile("./files/de-DE.ini")
	if err != nil {
		panic(err.Error())
	}
	return f
}

func TestIfLanguageExists(t *testing.T) {

	lang := "en-US"
	f := getIniFile()
	tp := make(pool)
	tp[lang] = f

	assert.True(t, tp.exists(lang), "en-US should exist in the pool.")
	assert.False(t, tp.exists("fr-FR"), "fr-FR should exist in the pool.")
}

func TestSaveLanguageIntoPool(t *testing.T) {

	lang := "en-US"
	tp := make(pool)
	err := tp.save(lang)
	assert.NoError(t, err, "It should find the ini file en-US.ini.")
	assert.True(t, tp.exists(lang), "en-US should exist in the pool.")

	err = tp.save("fr-FR")
	assert.Error(t, err, "fr-FR ini file does not exist.")
}

func TestReadLanguage(t *testing.T) {

	lang := "en-US"
	f := getIniFile()
	tp := make(pool)
	tp[lang] = f

	assert.Equal(t, tp.read(lang, "models/account", "text03"), "Password does not match security requirements.")
	assert.Panics(t, func() {
		tp.read(lang, "models/account", "foo")
	}, "The text does not exist.")

	assert.Equal(t, tp.read("fr-Fr", "models/account", "text03"), "Password does not match security requirements.")

}

func TestGermanLanguage(t *testing.T) {

	lang := "de-DE"
	f := getIniGermanFile()
	tp := make(pool)
	tp[lang] = f

	assert.Equal(t, tp.read(lang, "models/account", "text01"), "Der Name ist zu kurz.")

}

func TestTranslate(t *testing.T) {

	str := Translate("en-US", "models/account", "text04")
	assert.Equal(t, "Email address is not valid.", str)

	str = Translate("de-DE", "models/account", "text01")
	assert.Equal(t, "Der Name ist zu kurz.", str)

	str = Translate("fr-FR", "models/account", "text04")
	assert.Equal(t, "Email address is not valid.", str)

}
