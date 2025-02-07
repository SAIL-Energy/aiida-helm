package main

func main() {
	app := App{}
	app.Run(":8010")
}

/*
TODO
- Error handling, logging, tests
- Also install sealed secrets controller
- Add available chart versions
- Add REST access from AIIDA core (restricted access!)
- Restrict access to namespace
*/
