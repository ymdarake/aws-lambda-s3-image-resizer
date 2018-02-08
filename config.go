package main

import (
	"encoding/json"
	"regexp"
	"strings"
)

// Config ルールの配列を持つ、設定。
type Config struct {
	Rules []Rule
}

// Rule リサイズ対象となるか否かを判定するための正規表現 Path と、リサイズをする際の出力仕様を指定する OutputSpec(複数サイズに対応するため配列) を持つ。
type Rule struct {
	Path        string       `json:"path"`
	OutputSpecs []OutputSpec `json:"outputspecs"`
}

// OutputSpec リサイズする際の仕様。縦、横、出力先ディレクトリ(S3上)。
type OutputSpec struct {
	X         uint   `json:"x"`
	Y         uint   `json:"y"`
	Directory string `json:"directory"`
}

func ConfigureRules() *Config {
	return &Config{Rules: load()}
}

func (config *Config) ChooseRule(filename string) Rule {
	index := config.indexOf(filename)
	if index > -1 {
		return config.Rules[index]
	}
	return Rule{}
}

func (config *Config) indexOf(filename string) int {
	for index, r := range config.Rules {
		if regexp.MustCompile(r.Path).MatchString(strings.ToLower(filename)) {
			return index
		}
	}
	return -1
}

func load() []Rule {
	raw := RawRules()
	var r []Rule
	json.Unmarshal(raw, &r)
	return r
}
