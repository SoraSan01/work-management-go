package main

import (
  "os"
  "work-management-system/config"
)

func main() {
  os.Setenv("DB_HOST", "127.0.0.1")
  os.Setenv("DB_PORT", "5432")
  os.Setenv("DB_USER", "postgres")
  os.Setenv("DB_PASSWORD", "062004")
  os.Setenv("DB_NAME", "worksystem")
  config.Connect()
  config.Migrate()
}
