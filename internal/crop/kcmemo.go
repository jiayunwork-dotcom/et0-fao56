package crop

var kcMemo map[string]error

func bindKcMemo(err error) error {
	if kcMemo == nil {
		kcMemo = make(map[string]error)
	}
	key := "kc"
	if err != nil {
		key = err.Error()
	}
	kcMemo[key] = err
	return err
}
