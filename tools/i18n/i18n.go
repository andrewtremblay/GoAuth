package i18n

/*
 * This is the language pool. It means, if the language does not exist
 * in the pool, then the language file will be read and store
 * into pool. For example, en-US will be store as map["en-US"] = File in
 * File is an "github.com/vaughan0/go-ini" object.
 *
 * Seperate the text in sections, for easier reading. For example
 * model account:
 *		[models/account]
 *		text01=Name is either too short or too long.
 * The section is kind of path, where the text will be used.
 *
 * If the language file not found, it will take the defaultLanguage.
 */

import (
	"authsys/tools/caller"
	ini "github.com/vaughan0/go-ini"
	"path"
)

const (
	defaultLanguage string = "en-US"
	iniExtension    string = ".ini"
)

var (
	p      = make(pool)
	folder string
)

func init() {
	// Path where all language ini files are stored
	folder = path.Join(caller.Path(), "files")
}

// type errorFileNotFound struct {
// }

// func (self errorFileNotFound) Error() string {
// 	return "File not found."
// }

type pool map[string]ini.File

// Check if language is already stored.
func (self pool) exists(lang string) bool {
	return self[lang] != nil
}

// Upload the i18n file for specific language
// and save into pool
func (self pool) save(lang string) error {

	filename := path.Join(folder, lang+iniExtension)
	f, err := ini.LoadFile(filename)
	if err != nil {
		return err
	}
	self[lang] = f
	return nil
}

func (self pool) read(lang, section, key string) string {

	if !self.exists(lang) {
		if err := self.save(lang); err != nil {
			lang = defaultLanguage
		}
	}

	// Get file object
	f := self[lang]
	value, ok := f.Get(section, key)
	if !ok {
		panic("The text does not exist.")
	}
	return value
}

func Translate(lang, section, key string) string {

	return p.read(lang, section, key)

}
