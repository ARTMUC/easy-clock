# KidClock — TODO

## Status

- [x] #1  Napisz migracje SQL (001–007)
- [ ] #2  Zdefiniuj modele domenowe (internal/domain/)
- [ ] #3  Zdefiniuj interfejsy repozytoriów
- [ ] #4  Zaimplementuj repozytoria GORM (internal/infrastructure/persistence/)
- [ ] #5  Zaimplementuj serwisy aplikacyjne (internal/app/)
- [ ] #6  Zaimplementuj middleware auth JWT (internal/middleware/)
- [ ] #7  Zaimplementuj handlery HTTP i routing (internal/handler/ + internal/server/)
- [ ] #8  Zaimplementuj widok zegara (frontend — clock view)
- [ ] #9  Zaimplementuj konfigurator dla rodzica (frontend — configurator)
- [ ] #10 Skonfiguruj wiring aplikacji i docker-compose
- [ ] #11 Napisz testy integracyjne dla krytycznych ścieżek

## Zasady
- Po każdym tasku: code review + commit z jednolinijkowym message
- Przerwa na review przed kolejnym taskiem

## Tech Stack
- Go + Gin
- GORM + MySQL (zamiast sqlx z PRD — GORM już w projekcie)
- JWT (access token) + refresh token w DB
- Plain HTML + CSS + Vanilla JS (static files)
- Migracje: raw SQL w migrations/, uruchamiane ręcznie
