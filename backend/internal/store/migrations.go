package store

func (s *Store) Migrate() error {
	return s.db.AutoMigrate(&Job{}, &SubtitleAsset{})
}
