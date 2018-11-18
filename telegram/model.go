package telegram

type poll struct {
	pollName, owner string
	items           []string
}

type callbackData struct {
	PollName string `json:"poll_name"`
	Vote     string `json:"vote"`
}
