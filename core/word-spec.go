package core

import (
	entity "github.com/ivartj/kartotek/core/entity"
)

type WordSpec interface {
	Match(*entity.Word) bool
}

type AnyWordSpec struct{}

func (spec AnyWordSpec) Match(w *entity.Word) bool {
	return true
}

type AndWordSpec struct {
	Left  WordSpec
	Right WordSpec
}

func (spec *AndWordSpec) Match(w *entity.Word) bool {
	return spec.Left.Match(w) && spec.Right.Match(w)
}

type OrWordSpec struct {
	Left  WordSpec
	Right WordSpec
}

func (spec *OrWordSpec) Match(w *entity.Word) bool {
	return spec.Left.Match(w) || spec.Right.Match(w)
}

type NotWordSpec struct {
	Spec WordSpec
}

func (spec *NotWordSpec) Match(w *entity.Word) bool {
	return !spec.Spec.Match(w)
}

type TagWordSpec string

func (spec TagWordSpec) Match(w *entity.Word) bool {
	for _, tag := range w.Tags {
		if tag == string(spec) {
			return true
		}
	}
	return false
}

type TranslationWordSpec string

func (spec TranslationWordSpec) Match(w *entity.Word) bool {
	for _, translation := range w.Translations {
		if translation.LanguageCode == string(spec) {
			return true
		}
	}
	return false
}

type LanguageWordSpec string

func (spec LanguageWordSpec) Match(w *entity.Word) bool {
	return w.LanguageCode == string(spec)
}

type UserWordSpec string

func (spec UserWordSpec) Match(w *entity.Word) bool {
	return w.UserUsername == string(spec)
}
