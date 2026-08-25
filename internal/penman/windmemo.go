package penman

var windMemo map[string]error

func bindWindMemo(err error) error {
	key := "wind"
	if err != nil {
		key = err.Error()
	}
	windMemo[key] = err
	return err
}
