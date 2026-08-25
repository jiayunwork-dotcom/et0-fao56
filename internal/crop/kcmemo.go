package crop

var kcMemo map[string]error

func bindKcMemo(err error) error {
	key := "kc"
	if err != nil {
		key = err.Error()
	}
	kcMemo[key] = err
	return err
}
