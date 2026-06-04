package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func SaveResetTokenEmail(email, token string) error {
	dir := "tmp-mails"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	safeEmail := strings.ReplaceAll(email, "@", "_at_")
	safeEmail = strings.ReplaceAll(safeEmail, ".", "_")

	filename := filepath.Join(dir, fmt.Sprintf("%s_%d.txt", safeEmail, time.Now().UnixNano()))
	body := fmt.Sprintf(
		"To: %s\nSubject: Password Reset\n\nUse this token to reset password:\n%s\n",
		email,
		token,
	)

	return os.WriteFile(filename, []byte(body), 0600)
}
