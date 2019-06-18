/*
 * Copyright © 2019 Hedzr Yeh.
 */

package xs

// func LoadGwkConfig() {
// 	gwk.HeaderXServer = vxconf.GetBoolR("server.headerXServer")
//
// 	// fObj := vxconf.GetR("server.forwarders")
// 	// //fmt.Printf("Forwarders String Got: %v\n\n", fObj)
// 	//
// 	// b, err := yaml.Marshal(fObj)
// 	// check_panic(err)
// 	// s2 := string(b)
// 	// if gwk.Debug {
// 	// 	fmt.Printf("Forwarders Yaml Built: \nforwarders:\n%v\n\n", s2)
// 	// }
// 	// gwk.ThisConfig = gwk.DefaultForwardersConfig
// 	// gwk.ThisConfig.Forwarders = []*gwk.FwdItem{}
// 	// err = yaml.Unmarshal([]byte(s2), &gwk.ThisConfig.Forwarders)
// 	// check_panic(err)
// 	// //fmt.Printf("Forwarders Got #1: %v\n\n", thisConfig.Forwarders)
//
// 	// load the registrar configurations
// 	fObj := vxconf.GetR("server.registrar")
// 	// fmt.Printf("registrar String Got: %v\n\n", fObj)
//
// 	b, err := yaml.Marshal(fObj)
// 	check_panic(err)
// 	s2 := string(b)
// 	// logrus.Debugf("registrar Yaml Built: \n%v\n\n", s2)
// 	gwk.ThisConfig.Registrar = gwk.Registrar{}
// 	err = yaml.Unmarshal([]byte(s2), &gwk.ThisConfig.Registrar)
// 	check_panic(err)
//
// 	// Open the connection to registrar, enable `store`
// 	gwk.ThisConfig.Registrar.Open()
// }
//
// func check_panic(err error) {
// 	if err != nil {
// 		panic(err)
// 	}
// }
