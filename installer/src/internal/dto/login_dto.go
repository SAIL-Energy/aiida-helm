package dto

type LoginDto struct {
	UserId         string `json:"user_id,omitempty"`
	InstallerToken string `json:"installer_token,omitempty"`
}
