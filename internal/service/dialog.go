package service

import "github.com/wailsapp/wails/v3/pkg/application"

// DialogService menyediakan akses ke dialog native OS via Wails.
// Di-bind ke frontend sebagai RPC.
type DialogService struct{}

func NewDialogService() *DialogService { return &DialogService{} }

// PickDirectory membuka dialog pemilihan direktori native OS.
// Mengembalikan path yang dipilih, atau string kosong bila dibatalkan.
func (d *DialogService) PickDirectory() (string, error) {
	result, err := application.Get().Dialog.OpenFile().
		CanChooseDirectories(true).
		CanChooseFiles(false).
		SetTitle("Pilih direktori").
		PromptForSingleSelection()
	if err != nil {
		return "", err
	}
	return result, nil
}
