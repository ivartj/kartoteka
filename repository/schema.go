package repository

import (
	"github.com/ivartj/kartotek/core"
	"github.com/ivartj/kartotek/sqlmigrate"
)

const currentSchema = "ivartj-1"

func InitSchema(db core.DB) error {
	m, err := sqlmigrate.New(db)
	if err != nil {
		return err
	}
	err = m.RegisterMigration("", "ivartj-1", `

		create table user (
			user_id text not null
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
			image_id text not null
				primary key,
			mime_type text not null,
			license text not null,
			attribution text not null,
			attribution_url text not null
		);

		create table word (
			word_id text not null
				primary key,
			word text not null,
			language_code text not null
				references language(language_code),
			user_id text not null
				references user(user_id),
			image_id text
				-- can be null
				references image(image_id),
			notes text not null
		);

		create table word_translation (
			word_id text not null
				references word(word_id),
			language_code text not null
				references language(language_code),
			translation text not null
		);

		create table word_tag (
			word_id text not null
				references word(word_id),
			tag text not null
		);

		create view word_view as
		select
			word.*,
			group_concat(word_translation.language_code, ' ') as translation_codes,
			json_group_array(json_object(
				'word_id', json_quote(word_translation.word_id),
				'language_code', json_quote(word_translation.language_code),
				'translation', json_quote(word_translation.translation)
			)) filter ( where word_translation.language_code not null ) as translations,
			user.username
		from
			(select
					word.*,
					group_concat(word_tag.tag, ' ') as tags
				from
					word
					left outer join word_tag on word.word_id is word_tag.word_id
				group by word.word_id
			) word
			left outer join word_translation on word.word_id is word_translation.word_id
			natural join user
		group by word.word_id;
	`)
	if err != nil {
		return err
	}

	err = m.MigrateTo(currentSchema)
	if err != nil {
		return err
	}

	return nil
}
