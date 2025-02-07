package main

func main() {
	app := App{}
	app.Run(":8010")
}

/*
TODO
- Error handling, logging, tests
- Remove, upgrade, install AIIDA + sealed secrets controller
- Remove, upgrade, install other services
- Add REST access from AIIDA core (restricted access!)
- Restrict access to namespace
- Add possibility for additional values files
*/
