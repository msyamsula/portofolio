in /Users/m.syamsularifin/go/portofolio/backend-app/domain/agent/service/service.go

make the service dependent to /Users/m.syamsularifin/go/portofolio/backend-app/domain/agent/service/service.go

use interface as dependency so that domain is not directly tight to llm infrastructure

create new function that accept llm interface and return agent struct