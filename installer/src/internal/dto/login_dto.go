package dto

type LoginDto struct {
	UserId         string `json:"userId,omitempty"`
	InstallerToken string `json:"installerToken,omitempty"`
}
