package main

import "log"

func downloadAndInitialize(tfrVersion, tfVersion string) error {
	log.Printf("versions: terraformer %s and terraform %s\n", tfrVersion, tfVersion)
	log.Println("TODO: Download terraformer and terraform here, if not available")
	return nil
}
