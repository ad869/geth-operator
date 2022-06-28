package v1alpha1

// func TestCreateExtraDataFromValidators(t *testing.T) {
// 	list := []EthereumAddress{
// 		"0x8be3e1982e23f68d339cb551ade5b79f3dbdf648",
// 		"0x6e679cd21fe7d53b77ca284ecec7dac0f4ce78f6",
// 		"0x87864116c19de41f813709fcf159853fcc64496e",
// 		"0x8f5c7b47ea58387b7f593a10e760884cdbe56adf",
// 	}
// 	str, err := createExtraDataFromValidators(list)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	expert := "0xf87aa00000000000000000000000000000000000000000000000000000000000000000f854948be3e1982e23f68d339cb551ade5b79f3dbdf648946e679cd21fe7d53b77ca284ecec7dac0f4ce78f69487864116c19de41f813709fcf159853fcc64496e948f5c7b47ea58387b7f593a10e760884cdbe56adfc080c0"

// 	if str != expert {
// 		t.Error("is not equal")
// 		t.Log(str)
// 	}
// }
