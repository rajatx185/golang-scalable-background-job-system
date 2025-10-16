jobrunner/                    # module root (go.mod here)
├── go.mod
├── cmd/
│   ├── api/
│   │   └── main.go          # package main -> web API binary
│   └── worker/
│       └── main.go          # package main -> worker binary
├── internal/
│   ├── queue/               # internal packages (not importable outside module)
│   └── worker/
├── pkg/                     # optional: libraries intended for external use
│   └── metrics/
├── scripts/                 # dev scripts, db migrations, etc.
├── docker/                  # Docker assets
└── README.md
