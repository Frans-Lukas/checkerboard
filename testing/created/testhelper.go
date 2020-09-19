package created

import "log"

func failIf(err error, msg string) {
	if err != nil {
		log.Fatalf("error "+msg+": %v", err)
	}
}

func fatalFail(err error) {
	log.Fatalf("error: %v", err)
}
