package service

import "github.com/wailsapp/wails/v3/pkg/application"

type DialogService struct{}

func NewDialogService() *DialogService { return &DialogService{} }

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
