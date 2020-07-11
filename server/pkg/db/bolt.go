package db

func Get(bucketID, key []byte) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		tx.Bucket(secretBucket). .Put()
		return nil
	})
	if err != nil {
		log.Printf("error setting value: %v", err)
		http.Error(w, "error setting value", http.StatusInternalServerError)
		return
	}
}
