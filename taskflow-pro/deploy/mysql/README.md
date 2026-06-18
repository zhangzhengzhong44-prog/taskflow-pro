# MySQL Notes

Docker Compose creates the `taskflow_pro` database automatically through MySQL environment variables.

The backend runs GORM AutoMigrate on startup and creates these tables:

- `users`
- `projects`
- `project_members`
- `tasks`
- `comments`

