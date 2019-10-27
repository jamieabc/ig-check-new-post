package config_parser_test

import (
	"github.com/jamieabc/ig-check-new-post/config_parser"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	c, err := config_parser.Parse("fixture/test.yml")
	expected := config_parser.Config{Accounts: []string{
		"user1",
		"user2",
		"user3",
	}}

	assert.Equal(t, nil, err, "not empty error")
	assert.Equal(t, expected, c, "wrong config")
}
