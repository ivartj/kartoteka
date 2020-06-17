package repository

import (
	"github.com/ivartj/kartotek/core"
	"github.com/ivartj/kartotek/sqlmigrate"
)

func InitSchema(db core.DB) error {
	m, err := sqlmigrate.New(db)
	if err != nil {
		return err
	}
	err = m.RegisterMigration("", "ivartj-1", `

		create table user (
			user_id blob not null
				primary key,
			username text not null
				unique,
			email text
				-- can be null
				unique,
			email_unverified,
				-- can be null
			password_hash text not null
		);
		
		create table language (
			language_code text not null
				primary key,
			native_name text not null
		);

		create table image (
			image_id blob not null
				primary key,
			mime_type text not null,
			license text not null,
			attribution text not null,
			attribution_url text not null
		);

		create table word (
			word_id blob not null
				primary key,
			word text not null,
			language_code text not null
				references language(language_code),
			user_id blob not null
				references user(user_id),
			image_id blob
				-- can be null
				references image(image_id),
			notes text not null
		);

		create table word_translation (
			word_id blob not null
				references word(word_id),
			language_code text not null
				references language(language_code),
			translation text not null
		);

		create table word_tag (
			word_id blob not null
				references word(word_id),
			tag text not null
		);
	`)
	if err != nil {
		return err
	}

	err = m.MigrateTo("ivartj-1")
	if err != nil {
		return err
	}

	return nil
}
