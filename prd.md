# KidClock — Product Requirements Document

**Version:** 1.0  
**Date:** 2026-05-02  
**Status:** Draft

---

## 1. Purpose

KidClock is a web application that helps children understand the passage of time and the structure of their day. An analog clock face with activity rings visually shows the child what is happening now, what is coming up soon, and when important events are scheduled. A parent configures the schedule through an admin panel — the clock on the child's tablet or phone updates automatically.

---

## 2. Users

| Role | Description |
|---|---|
| **Parent** | Registers an account, adds children, configures profiles and the schedule |
| **Child** | Views the clock via a unique link — no login, no password required |

One parent account can manage multiple children. Sharing a child between two parent accounts (e.g. separated parents) is out of scope for MVP and planned for v2.

---

## 3. Tech Stack

```
Frontend:    plain HTML + CSS + Vanilla JS (served as static files by Go)
Backend:     Go (net/http or Echo/Gin)
DB layer:    sqlx + MySQL
Auth:        JWT (access token) + refresh token stored in DB
File upload: local disk → /static/uploads/
Migrations:  raw SQL, run manually
Timezone:    IANA timezone per child, resolved in Go via time.LoadLocation()
```

---

## 4. Project Structure

```
kidclock/
├── cmd/
│   └── server/
│       └── main.go                   # entry point, wiring
│
├── internal/
│   ├── domain/                       # pure domain models, no external dependencies
│   │   ├── user.go
│   │   ├── child.go
│   │   ├── profile.go                # Profile + Activity as inner entity
│   │   ├── schedule.go               # Schedule + DayAssignment
│   │   ├── event.go
│   │   └── clock.go                  # ClockState, ResolveActiveProfile logic
│   │
│   ├── app/                          # application layer — commands, orchestration
│   │   ├── user_service.go
│   │   ├── child_service.go
│   │   ├── profile_service.go
│   │   ├── schedule_service.go
│   │   ├── event_service.go
│   │   └── clock_service.go          # ResolveActiveProfile — core business logic
│   │
│   ├── persistence/                  # persistence layer — sqlx + repository adapters
│   │   ├── repository/
│   │   │   ├── user_repository.go
│   │   │   ├── child_repository.go
│   │   │   ├── profile_repository.go
│   │   │   ├── schedule_repository.go
│   │   │   └── event_repository.go
│   │   └── mysql.go                  # sqlx.Connect, *sqlx.DB init
│   │
│   ├── api/                          # HTTP handlers — request/response only, no business logic
│   │   ├── router.go
│   │   ├── middleware/
│   │   │   ├── auth.go               # JWT verification
│   │   │   └── cors.go
│   │   ├── handler/
│   │   │   ├── auth_handler.go
│   │   │   ├── child_handler.go
│   │   │   ├── profile_handler.go
│   │   │   ├── schedule_handler.go
│   │   │   ├── event_handler.go
│   │   │   └── clock_handler.go
│   │   └── dto/                      # request/response structs with json tags
│   │       ├── auth_dto.go
│   │       ├── child_dto.go
│   │       ├── profile_dto.go
│   │       ├── schedule_dto.go
│   │       ├── event_dto.go
│   │       └── clock_dto.go
│   │
│   └── upload/
│       └── storage.go                # file write to disk, returns path
│
├── migrations/                       # raw SQL, run manually
│   ├── 001_create_users.sql
│   ├── 002_create_children.sql
│   ├── 003_create_profiles.sql
│   ├── 004_create_activities.sql
│   ├── 005_create_schedules.sql
│   ├── 006_create_events.sql
│   ├── 007_seed_preset_activities.sql
│   └── README.md                     # how to run: mysql -u root kidclock < migrations/001_...sql
│
├── static/
│   ├── index.html
│   ├── app.js
│   ├── presets/                      # bundled preset activity images
│   └── uploads/                      # user-uploaded images
│
├── go.mod
├── go.sum
└── .env                              # DSN, JWT_SECRET, PORT, etc.
```

**Dependency rule:** `internal/domain/` imports nothing from `persistence/` or `api/`. Dependencies point inward only. Services in `app/` depend on repository interfaces (defined in `domain/` or `app/`), not on concrete sqlx implementations — making them easy to mock in tests.

---

## 5. Data Model

### 5.1 Tables

```sql
-- users
CREATE TABLE users (
  id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  email         VARCHAR(255) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- refresh_tokens
CREATE TABLE refresh_tokens (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id    BIGINT UNSIGNED NOT NULL,
  token_hash VARCHAR(255) NOT NULL,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- children
CREATE TABLE children (
  id                 BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  user_id            BIGINT UNSIGNED NOT NULL,
  name               VARCHAR(100) NOT NULL,
  timezone           VARCHAR(100) NOT NULL DEFAULT 'Europe/Warsaw',
  avatar_path        VARCHAR(500),
  default_profile_id BIGINT UNSIGNED,
  clock_token        CHAR(64) NOT NULL UNIQUE,  -- random hex, generated on child creation, immutable
  created_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- profiles
CREATE TABLE profiles (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  child_id   BIGINT UNSIGNED NOT NULL,
  name       VARCHAR(100) NOT NULL,
  color      VARCHAR(20) NOT NULL DEFAULT '#e07a3a',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE
);

-- activities
CREATE TABLE activities (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  profile_id  BIGINT UNSIGNED NOT NULL,
  preset_id   BIGINT UNSIGNED DEFAULT NULL,     -- nullable FK to preset_activities
  emoji       VARCHAR(10) NOT NULL,
  label       VARCHAR(100) NOT NULL,
  from_hour   TINYINT UNSIGNED NOT NULL,
  to_hour     TINYINT UNSIGNED NOT NULL,
  ring        TINYINT UNSIGNED NOT NULL,        -- 1=AM, 2=PM
  image_path  VARCHAR(500) NOT NULL,            -- required; points to presets/ or uploads/
  sort_order  SMALLINT UNSIGNED NOT NULL DEFAULT 0,
  FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE,
  FOREIGN KEY (preset_id) REFERENCES preset_activities(id) ON DELETE SET NULL
);

-- preset_activities — system-wide, read-only, bundled with the app
CREATE TABLE preset_activities (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  emoji      VARCHAR(10) NOT NULL,
  label      VARCHAR(100) NOT NULL,
  image_path VARCHAR(500) NOT NULL,             -- e.g. /static/presets/sleep.png
  ring       TINYINT UNSIGNED NOT NULL,         -- suggested ring (1=AM, 2=PM), overridable
  sort_order SMALLINT UNSIGNED NOT NULL DEFAULT 0
);

-- schedule_days (one row per day of the week per child)
CREATE TABLE schedule_days (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  child_id    BIGINT UNSIGNED NOT NULL,
  day_of_week TINYINT UNSIGNED NOT NULL,        -- 0=Sunday, 1=Mon, ..., 6=Sat
  profile_id  BIGINT UNSIGNED NOT NULL,
  UNIQUE KEY uq_child_day (child_id, day_of_week),
  FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE,
  FOREIGN KEY (profile_id) REFERENCES profiles(id)
);

-- events (one-off overrides)
CREATE TABLE events (
  id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  child_id    BIGINT UNSIGNED NOT NULL,
  date        DATE NOT NULL,
  from_time   TIME NOT NULL,
  to_time     TIME NOT NULL,
  label       VARCHAR(200) NOT NULL,
  emoji       VARCHAR(10),
  image_path  VARCHAR(500),
  profile_id  BIGINT UNSIGNED,                  -- optional: reference to an existing profile
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (child_id) REFERENCES children(id) ON DELETE CASCADE,
  FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE SET NULL
);

-- event_activities (inline activities for events that don't use a profile)
CREATE TABLE event_activities (
  id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  event_id   BIGINT UNSIGNED NOT NULL,
  emoji      VARCHAR(10) NOT NULL,
  label      VARCHAR(100) NOT NULL,
  from_hour  TINYINT UNSIGNED NOT NULL,
  to_hour    TINYINT UNSIGNED NOT NULL,
  ring       TINYINT UNSIGNED NOT NULL,
  image_path VARCHAR(500),
  FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE CASCADE
);
```

---

### 5.2 Business Rules

- `activities.from_hour < activities.to_hour`, both in range 0–23
- Activities within one profile and one ring must not overlap in time
- `events.from_time < events.to_time`
- An event uses either `profile_id` or `event_activities` — not both simultaneously (enforced in service layer)
- `children.clock_token` — 64-character random hex, generated on child creation, never changes
- `activities.image_path` is required — either from a preset or uploaded by the parent

---

### 5.3 Preset Activities

The system ships with a set of built-in activities with default images. The parent only needs to set the hours. If a parent wants to create a custom activity (e.g. "Dentist visit"), they must upload their own image.

#### Initial preset seed

| Emoji | Label | Suggested ring | File |
|---|---|---|---|
| ☀️ | Wake up | AM | presets/wake_up.png |
| 🥣 | Breakfast | AM | presets/breakfast.png |
| 🦷 | Brush teeth | AM | presets/teeth.png |
| 🎒 | School | AM | presets/school.png |
| 📚 | Homework | AM | presets/homework.png |
| 🍽️ | Lunch | PM | presets/lunch.png |
| ⚽ | Sport / outdoor play | PM | presets/sport.png |
| 🎨 | Creative time | PM | presets/creative.png |
| 🎮 | Free time / screen time | PM | presets/freetime.png |
| 🛁 | Bath | PM | presets/bath.png |
| 🌙 | Wind down | PM | presets/wind_down.png |
| 🛏️ | Sleep | PM | presets/sleep.png |

Seed is applied once via migration `007_seed_preset_activities.sql`.

#### Rules

- Presets are **global** — shared across all users, read-only
- When adding an activity to a profile, the parent sees two tabs: **"Presets"** (list of built-in activities) and **"Custom"** (form with image upload field)
- Selecting a preset copies `emoji`, `label`, `image_path`, and the suggested `ring` into the new activity — the parent sets only `from_hour` and `to_hour`
- A custom activity requires uploading an image — `image_path` is mandatory (validated in service layer)
- The parent may change the `label` and `emoji` of a preset-based activity; `image_path` remains a reference to `/static/presets/` — the file is not physically copied
- The nullable `preset_id` FK in `activities` distinguishes preset-based entries from custom ones and enables future global image updates for presets

---

## 6. API Endpoints

### Auth
```
POST   /api/auth/register          RegisterUser
POST   /api/auth/login             LoginUser  → returns access_token + refresh_token
POST   /api/auth/refresh           refresh access_token
POST   /api/auth/logout            revoke refresh_token
```

### Children
```
GET    /api/children               list children of logged-in parent
POST   /api/children               AddChild
GET    /api/children/:id           child details
PUT    /api/children/:id           UpdateChild
DELETE /api/children/:id           RemoveChild
POST   /api/children/:id/avatar    upload avatar
```

### Profiles
```
GET    /api/children/:id/profiles         list profiles for a child
POST   /api/children/:id/profiles         CreateProfile
GET    /api/profiles/:id                  profile details with activities
PUT    /api/profiles/:id                  UpdateProfile
DELETE /api/profiles/:id                  DeleteProfile
PUT    /api/children/:id/default-profile  SetDefaultProfile

POST   /api/profiles/:id/activities       AddActivityToProfile
PUT    /api/activities/:id                UpdateActivity
DELETE /api/activities/:id                RemoveActivity
POST   /api/activities/:id/image          UploadActivityImage
```

### Schedule
```
GET    /api/children/:id/schedule         weekly schedule (7 days)
PUT    /api/children/:id/schedule         bulk upsert — AssignProfileToDays
DELETE /api/children/:id/schedule/:day    ClearDayAssignment
```

### Events
```
GET    /api/children/:id/events           list events (params: ?from=&to=)
POST   /api/children/:id/events           CreateEvent
PUT    /api/events/:id                    UpdateEvent
DELETE /api/events/:id                    DeleteEvent
```

### Clock (public — auth via clock_token)
```
GET    /api/clock/:clock_token            GetClockState → active profile + active activity
```

### Presets (public)
```
GET    /api/preset-activities             list all preset activities
```

### Upload
```
POST   /api/upload                        upload image file → returns image_path
```

---

## 7. ResolveActiveProfile Logic

The core of the system — called by `GET /api/clock/:clock_token`:

```go
func (s *ClockService) Resolve(ctx context.Context, child *domain.Child, now time.Time) (*domain.ClockState, error) {
    loc, _ := time.LoadLocation(child.Timezone)
    localNow := now.In(loc)

    // 1. Check one-off events — highest priority
    events, _ := s.eventRepo.GetForDate(ctx, child.ID, localNow.Format("2006-01-02"))
    for _, e := range events {
        if timeInRange(localNow, e.FromTime, e.ToTime) {
            return resolveFromEvent(e, localNow), nil
        }
    }

    // 2. Check weekly schedule
    weekday := int(localNow.Weekday())
    if assignment, ok := s.scheduleRepo.GetDay(ctx, child.ID, weekday); ok {
        profile, _ := s.profileRepo.GetWithActivities(ctx, assignment.ProfileID)
        return resolveFromProfile(profile, localNow), nil
    }

    // 3. Fallback — default profile
    if child.DefaultProfileID != nil {
        profile, _ := s.profileRepo.GetWithActivities(ctx, *child.DefaultProfileID)
        return resolveFromProfile(profile, localNow), nil
    }

    return &domain.ClockState{Empty: true}, nil
}
```

Priority: **Event > DayAssignment > DefaultProfile**

---

## 8. Event Storming

### 8.1 Domain Events

**Auth / User**
- `UserRegistered`
- `UserLoggedIn`
- `UserLoggedOut`
- `PasswordChanged`
- `RefreshTokenIssued`
- `RefreshTokenRevoked`

**Children**
- `ChildAdded`
- `ChildUpdated`
- `ChildRemoved`
- `ChildAvatarUploaded`
- `DefaultProfileSet`

**Profiles**
- `ProfileCreated`
- `ProfileUpdated`
- `ProfileDeleted`
- `ActivityAddedToProfile`
- `ActivityUpdatedInProfile`
- `ActivityRemovedFromProfile`
- `ActivityImageUploaded`

**Schedule**
- `WeeklyScheduleSet` (one or multiple days at once)
- `WeeklyScheduleCleared`

**Events**
- `EventCreated`
- `EventUpdated`
- `EventDeleted`
- `EventOccurred` ← runtime: clock entered the event's time window

**Clock / Runtime**
- `ClockViewed`
- `ActiveProfileResolved` ← system selected a profile (event > schedule > default)
- `ActiveActivityResolved` ← which activity is currently active on the ring

---

### 8.2 Commands

**Auth / User**
- `RegisterUser(email, password)`
- `LoginUser(email, password)`
- `LogoutUser(refresh_token)`
- `ChangePassword(old, new)`
- `RefreshAccessToken(refresh_token)`

**Children**
- `AddChild(name, timezone, avatar?)`
- `UpdateChild(name, timezone, avatar?)`
- `RemoveChild(child_id)`
- `SetDefaultProfile(child_id, profile_id)`

**Profiles**
- `CreateProfile(child_id, name, color)`
- `UpdateProfile(profile_id, name, color)`
- `DeleteProfile(profile_id)`
- `AddActivityToProfile(profile_id, emoji, label, from_hour, to_hour, ring, preset_id? | image_file)`
- `UpdateActivity(activity_id, ...)`
- `RemoveActivity(activity_id)`
- `UploadActivityImage(activity_id, file)`

**Schedule**
- `AssignProfileToDay(child_id, day_of_week, profile_id)`
- `AssignProfileToDays(child_id, days[], profile_id)` ← bulk, e.g. Mon–Fri in one action
- `ClearDayAssignment(child_id, day_of_week)`

**Events**
- `CreateEvent(child_id, date, from_time, to_time, label, emoji, profile_id | activities[])`
- `UpdateEvent(event_id, ...)`
- `DeleteEvent(event_id)`

**Clock / Runtime**
- `ResolveActiveProfile(clock_token, datetime)`
- `GetClockState(clock_token, datetime)`

---

### 8.3 Aggregates

**User**
- Root: `User { id, email, password_hash, created_at }`
- Invariants: unique email, password min. 8 characters

**Child**
- Root: `Child { id, user_id, name, timezone, avatar_path, default_profile_id, clock_token }`
- Invariants: valid IANA timezone, clock_token is unique and immutable, default_profile must belong to this child

**Profile**
- Root: `Profile { id, child_id, name, color }`
- Inner entities: `Activity { id, emoji, label, from_hour, to_hour, ring, image_path, preset_id? }`
- Invariants: activities do not overlap within the same ring, from_hour < to_hour, cannot delete a profile assigned to a schedule day or event

**Schedule**
- Root: `Schedule { child_id }` ← singleton per child
- Inner entities: `DayAssignment { day_of_week, profile_id }`
- Invariants: one day = at most one profile, profile must belong to the same child

**Event**
- Root: `Event { id, child_id, date, from_time, to_time, label, emoji, image_path }`
- Value objects: `profile_id` (reference) XOR `EventActivities[]` (inline) — not both
- Invariants: from_time < to_time, referenced profile_id must belong to the child, warning (not hard block) on overlapping events

**ClockState** ← read model, not an aggregate
- Projection from Event + Schedule + Profile
- Priority: Event > DayAssignment > DefaultProfile

---

## 9. Application Views

### 9.1 Clock View (`/clock/:clock_token`)

- Fullscreen canvas 360×360px
- Polls `/api/clock/:token` every 60 seconds
- Two activity rings: AM (outer), PM (inner)
- Active ring: 69px wide; inactive ring: 28px wide
- Dual hour numbers (e.g. `9` and `21`) side by side on the outer ring — active time is displayed larger
- Spokes from the number ring to the center (12 lines, major ones at 3h intervals are thicker)
- Minute hand reaches to the number ring; hour hand is shorter
- Light background (AM) / dark background (PM)
- No navigation, no back button

### 9.2 Configurator (`/app/config`)

Requires login. Sections:

**Children** — list, add, edit (name, timezone, avatar), "Copy clock link" button

**Profiles** — list of profiles per child, create/edit:
- name, color
- list of activities with emoji, label, from–to hours, ring (AM/PM), image
- two tabs when adding: **"Presets"** (pick from built-in list, set hours only) and **"Custom"** (full form, image upload required)

**Weekly Schedule** — 7 days of the week, per-day profile selector (or "none")

**Events** — list of upcoming events, form: date, from–to time, label, emoji, choose profile or define inline activities

---

## 10. Clock Mockup — HTML/CSS/JS

Complete working mockup developed during design phase:

```html
<!DOCTYPE html>
<html lang="pl">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Zegar Dziecka</title>
<link href="https://fonts.googleapis.com/css2?family=Nunito:wght@400;500;600;700&family=Baloo+2:wght@500;700&display=swap" rel="stylesheet">
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

  :root {
    --bg: #f5f0e8;
    --surface: #fffdf8;
    --border: rgba(0,0,0,0.08);
    --text: #2c2416;
    --text-muted: #8a7a62;
    --accent: #e07a3a;
    --accent-light: #fdf0e6;
    --radius: 16px;
    --shadow: 0 2px 12px rgba(0,0,0,0.08);
  }

  body {
    font-family: 'Nunito', sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 32px 16px;
  }

  h1 { font-family: 'Baloo 2', sans-serif; font-size: 28px; font-weight: 700;
       color: var(--text); margin-bottom: 28px; letter-spacing: -0.5px; }
  h1 span { color: var(--accent); }

  #layout { display: flex; gap: 28px; align-items: flex-start; width: 100%;
             max-width: 860px; flex-wrap: wrap; justify-content: center; }
  #clock-wrap { flex: 0 0 auto; display: flex; flex-direction: column;
                align-items: center; gap: 12px; }
  canvas { display: block; filter: drop-shadow(0 4px 20px rgba(0,0,0,0.12)); }
  #clock-note { font-size: 12px; color: var(--text-muted); }
  #panel { flex: 1; min-width: 280px; max-width: 360px;
           display: flex; flex-direction: column; gap: 14px; }
  .card { background: var(--surface); border: 1px solid var(--border);
          border-radius: var(--radius); padding: 14px 16px; box-shadow: var(--shadow); }
  .card-label { font-family: 'Baloo 2', sans-serif; font-size: 11px; font-weight: 500;
                letter-spacing: 0.09em; text-transform: uppercase;
                color: var(--text-muted); margin-bottom: 10px; }
  .time-row { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
  .time-row label { font-size: 13px; color: var(--text-muted); }
  input[type=number], input[type=text] {
    font-family: 'Nunito', sans-serif; font-size: 13px;
    border: 1px solid var(--border); border-radius: 8px; padding: 5px 8px;
    background: var(--bg); color: var(--text); outline: none; transition: border-color 0.15s;
  }
  input[type=number]:focus, input[type=text]:focus { border-color: var(--accent); }
  input[type=number] { width: 58px; }
  input[type=text] { flex: 1; min-width: 70px; }
  .now-btn, .add-btn {
    font-family: 'Nunito', sans-serif; font-size: 12px; font-weight: 600;
    cursor: pointer; border-radius: 8px; border: 1px solid var(--border);
    background: var(--bg); color: var(--text-muted); padding: 5px 12px;
    transition: background 0.15s, color 0.15s;
  }
  .now-btn:hover, .add-btn:hover {
    background: var(--accent-light); color: var(--accent); border-color: var(--accent);
  }
  .add-btn { width: 100%; margin-top: 6px; text-align: center; padding: 7px; }
  .activity-row { display: flex; align-items: center; gap: 6px; margin-bottom: 8px; flex-wrap: wrap; }
  .activity-row:last-of-type { margin-bottom: 0; }
  .activity-row label { font-size: 11px; color: var(--text-muted); }
  .emoji-btn {
    font-size: 16px; border: 1px solid var(--border); border-radius: 8px;
    width: 32px; height: 30px; display: flex; align-items: center;
    justify-content: center; background: var(--bg); cursor: pointer;
    flex-shrink: 0; transition: border-color 0.15s;
  }
  .emoji-btn:hover { border-color: var(--accent); }
  .remove-btn {
    background: none; border: none; font-size: 16px; cursor: pointer;
    color: var(--text-muted); line-height: 1; padding: 2px 4px;
    flex-shrink: 0; border-radius: 6px; transition: background 0.15s;
  }
  .remove-btn:hover { background: #fee2e2; color: #b91c1c; }
  #upload-area {
    border: 1.5px dashed rgba(0,0,0,0.15); border-radius: 10px; padding: 12px;
    text-align: center; cursor: pointer; font-size: 12px; color: var(--text-muted);
    transition: border-color 0.15s, background 0.15s;
  }
  #upload-area:hover { border-color: var(--accent); background: var(--accent-light); }
  #upload-area input { display: none; }
  #img-preview { max-height: 48px; border-radius: 8px; display: none; margin-bottom: 4px; }
  #img-time-config { display: none; margin-top: 10px; }
  .emoji-picker {
    display: none; flex-wrap: wrap; gap: 4px; padding: 8px;
    background: var(--surface); border-radius: 10px; border: 1px solid var(--border);
    box-shadow: 0 4px 16px rgba(0,0,0,0.1); position: fixed; z-index: 100; max-width: 240px;
  }
  .emoji-picker.open { display: flex; }
  .emoji-picker span { font-size: 20px; cursor: pointer; padding: 3px 4px;
                       border-radius: 6px; transition: background 0.1s; }
  .emoji-picker span:hover { background: var(--bg); }
  #ampm-badge {
    display: inline-flex; align-items: center; gap: 8px;
    background: var(--surface); border: 1px solid var(--border);
    border-radius: 100px; padding: 5px 14px;
    font-family: 'Baloo 2', sans-serif; font-size: 13px; font-weight: 500;
    color: var(--text-muted); box-shadow: var(--shadow);
  }
  #ampm-badge .dot { width: 8px; height: 8px; border-radius: 50%; background: var(--accent); }
</style>
</head>
<body>

<h1>Zegar <span>dla dziecka</span></h1>
<div id="ampm-badge">
  <span class="dot" id="badge-dot"></span>
  <span id="badge-text">Przed południem</span>
</div>
<br>

<div id="layout">
  <div id="clock-wrap">
    <canvas id="clockCanvas" width="360" height="360"></canvas>
    <div id="clock-note">Zmień godzinę w panelu → lub kliknij "Teraz"</div>
  </div>
  <div id="panel">
    <div class="card">
      <div class="card-label">Aktualny czas</div>
      <div class="time-row">
        <label>Godz.</label>
        <input type="number" id="simH" min="0" max="23" value="10" onchange="updateSim()">
        <label>Min.</label>
        <input type="number" id="simM" min="0" max="59" value="0" onchange="updateSim()">
        <button class="now-btn" onclick="setNow()">Teraz</button>
      </div>
    </div>
    <div class="card">
      <div class="card-label">Pierścień AM — przed południem</div>
      <div id="ring1-items"></div>
      <button class="add-btn" onclick="addActivity(1)">+ Dodaj aktywność</button>
    </div>
    <div class="card">
      <div class="card-label">Pierścień PM — po południu</div>
      <div id="ring2-items"></div>
      <button class="add-btn" onclick="addActivity(2)">+ Dodaj aktywność</button>
    </div>
    <div class="card">
      <div class="card-label">Własny obrazek</div>
      <div id="upload-area" onclick="document.getElementById('imgInput').click()">
        <img id="img-preview" alt="podgląd">
        <div id="upload-label">📁 Kliknij, aby wgrać obrazek</div>
        <input type="file" id="imgInput" accept="image/*" onchange="loadImage(event)">
      </div>
      <div id="img-time-config">
        <div class="time-row" style="flex-wrap:wrap;gap:8px;margin-top:4px;">
          <label>Od godz.</label><input type="number" id="imgFrom" min="0" max="23" value="7" onchange="draw()">
          <label>Do godz.</label><input type="number" id="imgTo" min="0" max="23" value="19" onchange="draw()">
          <label>Pierścień</label><input type="number" id="imgRing" min="1" max="2" value="1" onchange="draw()">
        </div>
      </div>
    </div>
  </div>
</div>

<div id="emoji-popup" class="emoji-picker">
  <span>🌙</span><span>⭐</span><span>☀️</span><span>🌤️</span><span>🍽️</span><span>🥣</span><span>🥛</span><span>🍎</span>
  <span>🎒</span><span>📚</span><span>✏️</span><span>🖊️</span><span>🎨</span><span>🎮</span><span>🧸</span><span>⚽</span>
  <span>🚲</span><span>🏃</span><span>🛁</span><span>🦷</span><span>🛏️</span><span>💤</span><span>📖</span><span>🎵</span>
</div>

<script>
const canvas = document.getElementById('clockCanvas');
const ctx = canvas.getContext('2d');
const CX = 180, CY = 180, R = 175;

let simHour = 10, simMinute = 0, customImg = null, activeEmojiTarget = null;

let activities = {
  1: [
    {emoji:'☀️', label:'Wstaj',     from:6,  to:8},
    {emoji:'🥣', label:'Śniadanie', from:8,  to:9},
    {emoji:'📚', label:'Nauka',     from:9,  to:12}
  ],
  2: [
    {emoji:'🍽️', label:'Obiad',  from:12, to:13},
    {emoji:'🎮', label:'Zabawa', from:14, to:17},
    {emoji:'🛁', label:'Kąpiel', from:18, to:19},
    {emoji:'🛏️', label:'Sen',   from:20, to:22}
  ]
};

const W_THICK = 69, W_THIN = 28, GAP = 3, NUM_R = R - 16;

function getRingParams(h) {
  const isAM = h < 12;
  const w1 = isAM ? W_THICK : W_THIN;
  const w2 = isAM ? W_THIN  : W_THICK;
  const ring1Outer = NUM_R - 13;
  const r1 = ring1Outer - w1 / 2;
  const r2 = ring1Outer - w1 - GAP - w2 / 2;
  const centerR = ring1Outer - w1 - GAP - w2 - GAP;
  return { isAM, w1, w2, r1, r2, centerR };
}

function setNow() {
  const d = new Date();
  simHour = d.getHours(); simMinute = d.getMinutes();
  document.getElementById('simH').value = simHour;
  document.getElementById('simM').value = simMinute;
  draw();
}
function updateSim() {
  simHour = parseInt(document.getElementById('simH').value)||0;
  simMinute = parseInt(document.getElementById('simM').value)||0;
  draw();
}
function loadImage(e) {
  const file = e.target.files[0]; if (!file) return;
  const reader = new FileReader();
  reader.onload = ev => {
    const img = new Image();
    img.onload = () => { customImg = img; draw(); };
    img.src = ev.target.result;
    document.getElementById('img-preview').src = ev.target.result;
    document.getElementById('img-preview').style.display = 'block';
    document.getElementById('upload-label').textContent = file.name;
    document.getElementById('img-time-config').style.display = 'block';
  };
  reader.readAsDataURL(file);
}
function addActivity(ring) {
  activities[ring].push({emoji:'⭐', label:'Aktywność', from:ring===1?8:13, to:ring===1?9:14});
  renderLists(); draw();
}
function removeActivity(ring, i) { activities[ring].splice(i,1); renderLists(); draw(); }

function renderLists() {
  [1,2].forEach(ring => {
    const el = document.getElementById(`ring${ring}-items`);
    el.innerHTML = '';
    activities[ring].forEach((a,i) => {
      const row = document.createElement('div');
      row.className = 'activity-row';
      row.innerHTML = `
        <button class="emoji-btn" onclick="openEmojiPicker(${ring},${i},this)">${a.emoji}</button>
        <input type="text" value="${a.label}" oninput="activities[${ring}][${i}].label=this.value;draw()">
        <label>od</label>
        <input type="number" value="${a.from}" min="0" max="23" style="width:52px" onchange="activities[${ring}][${i}].from=+this.value;draw()">
        <label>do</label>
        <input type="number" value="${a.to}" min="0" max="23" style="width:52px" onchange="activities[${ring}][${i}].to=+this.value;draw()">
        <button class="remove-btn" onclick="removeActivity(${ring},${i})">×</button>`;
      el.appendChild(row);
    });
  });
}

function openEmojiPicker(ring, idx, btn) {
  activeEmojiTarget = {ring, idx};
  const popup = document.getElementById('emoji-popup');
  const rect = btn.getBoundingClientRect();
  popup.style.top = (rect.bottom + window.scrollY + 4) + 'px';
  popup.style.left = rect.left + 'px';
  popup.classList.toggle('open');
}
document.getElementById('emoji-popup').querySelectorAll('span').forEach(sp => {
  sp.addEventListener('click', () => {
    if (!activeEmojiTarget) return;
    activities[activeEmojiTarget.ring][activeEmojiTarget.idx].emoji = sp.textContent;
    document.getElementById('emoji-popup').classList.remove('open');
    renderLists(); draw();
  });
});
document.addEventListener('click', e => {
  if (!e.target.closest('.emoji-btn') && !e.target.closest('#emoji-popup'))
    document.getElementById('emoji-popup').classList.remove('open');
});

function hsl(h,s,l,a=1){ return `hsla(${h},${s}%,${l}%,${a})`; }
function hourToAngle(h){ return (h/12)*Math.PI*2 - Math.PI/2; }

function drawSpokes(centerR, spokeOuterR, isAM) {
  for (let i = 0; i < 12; i++) {
    const isMajor = i % 3 === 0;
    const angle = hourToAngle(i === 0 ? 12 : i);
    const x1 = CX + spokeOuterR * Math.cos(angle);
    const y1 = CY + spokeOuterR * Math.sin(angle);
    const x2 = CX + (centerR + 1) * Math.cos(angle);
    const y2 = CY + (centerR + 1) * Math.sin(angle);
    ctx.beginPath(); ctx.moveTo(x1, y1); ctx.lineTo(x2, y2);
    ctx.strokeStyle = isAM
      ? (isMajor ? hsl(35,50,40,0.32) : hsl(35,40,50,0.18))
      : (isMajor ? hsl(220,50,75,0.32) : hsl(220,40,70,0.18));
    ctx.lineWidth = isMajor ? 1 : 0.5;
    ctx.lineCap = 'round'; ctx.stroke();
  }
}

function drawRingActivities(list, ringR, ringW, opacity, currentH, ringIdx) {
  const colorHues = [200,35,130,280,350,175,55,320];
  const imgFrom  = parseInt(document.getElementById('imgFrom').value)||0;
  const imgTo    = parseInt(document.getElementById('imgTo').value)||0;
  const imgRingV = parseInt(document.getElementById('imgRing').value)||1;

  list.forEach((a,i) => {
    let from12 = a.from % 12, to12 = a.to % 12;
    if (to12===0 && a.to!==0) to12=12;
    if (from12===0 && a.from!==0) from12=12;
    const startA = hourToAngle(from12);
    let endA = hourToAngle(to12);
    if (endA <= startA) endA += Math.PI*2;
    const active = currentH >= a.from && currentH < a.to;
    const hue = colorHues[i % colorHues.length];

    ctx.beginPath(); ctx.arc(CX, CY, ringR, startA, endA);
    ctx.strokeStyle = hsl(hue, 65, 60, opacity * (active ? 1 : 0.35));
    ctx.lineWidth = ringW; ctx.lineCap = 'butt'; ctx.stroke();

    if (ringW < 10) return;
    const midA = (startA + endA) / 2;
    const ex = CX + ringR * Math.cos(midA);
    const ey = CY + ringR * Math.sin(midA);

    if (customImg && imgRingV === ringIdx) {
      const imgActive = currentH >= imgFrom && currentH < imgTo;
      const isize = Math.min(ringW * 0.88, 62);
      ctx.save(); ctx.globalAlpha = opacity * (imgActive ? 1 : 0.2);
      ctx.beginPath(); ctx.arc(ex, ey, isize/2, 0, Math.PI*2); ctx.clip();
      ctx.drawImage(customImg, ex-isize/2, ey-isize/2, isize, isize);
      ctx.restore();
    } else {
      const emojiSize = Math.min(ringW * 0.68, 44);
      ctx.save(); ctx.globalAlpha = opacity * (active ? 1 : 0.35);
      ctx.font = `${emojiSize}px serif`;
      ctx.textAlign = 'center'; ctx.textBaseline = 'middle';
      ctx.fillText(a.emoji, ex, ey); ctx.restore();
    }
  });
}

function draw() {
  ctx.clearRect(0, 0, canvas.width, canvas.height);
  const h = simHour, m = simMinute;
  const { isAM, w1, w2, r1, r2, centerR } = getRingParams(h);

  document.getElementById('badge-text').textContent = isAM ? 'Przed południem' : 'Po południu';
  document.getElementById('badge-dot').style.background = isAM ? '#e07a3a' : '#4a7abf';

  ctx.beginPath(); ctx.arc(CX,CY,R,0,Math.PI*2);
  ctx.fillStyle = isAM ? hsl(48,90,96) : hsl(228,40,10); ctx.fill();
  ctx.beginPath(); ctx.arc(CX,CY,R-1,0,Math.PI*2);
  ctx.strokeStyle = isAM ? hsl(45,55,75,0.5) : hsl(228,35,30,0.6);
  ctx.lineWidth=2; ctx.stroke();

  ctx.textAlign='center'; ctx.textBaseline='middle';
  const numPairs = [
    {h12:12,am:'12',pm:'0'},{h12:1,am:'1',pm:'13'},{h12:2,am:'2',pm:'14'},
    {h12:3,am:'3',pm:'15'},{h12:4,am:'4',pm:'16'},{h12:5,am:'5',pm:'17'},
    {h12:6,am:'6',pm:'18'},{h12:7,am:'7',pm:'19'},{h12:8,am:'8',pm:'20'},
    {h12:9,am:'9',pm:'21'},{h12:10,am:'10',pm:'22'},{h12:11,am:'11',pm:'23'},
  ];
  numPairs.forEach(p => {
    const angle = hourToAngle(p.h12);
    const nx = CX + NUM_R * Math.cos(angle);
    const ny = CY + NUM_R * Math.sin(angle);
    const tx = -Math.sin(angle), ty = Math.cos(angle), off = 7;
    ctx.font = `${isAM?'700':'400'} ${isAM?14:8}px 'Nunito', sans-serif`;
    ctx.fillStyle = isAM ? hsl(35,60,28) : hsl(40,25,60,0.38);
    ctx.fillText(p.am, nx - tx*off, ny - ty*off);
    ctx.font = `${isAM?'400':'700'} ${isAM?8:14}px 'Nunito', sans-serif`;
    ctx.fillStyle = isAM ? hsl(220,45,50,0.38) : hsl(220,50,84);
    ctx.fillText(p.pm, nx + tx*off, ny + ty*off);
  });

  const tickOuter = NUM_R - 12;
  for (let i=0;i<12;i++) {
    const a = hourToAngle(i===0?12:i);
    const len = i%3===0?9:5;
    ctx.beginPath();
    ctx.moveTo(CX+tickOuter*Math.cos(a), CY+tickOuter*Math.sin(a));
    ctx.lineTo(CX+(tickOuter-len)*Math.cos(a), CY+(tickOuter-len)*Math.sin(a));
    ctx.strokeStyle = isAM ? hsl(35,30,55,0.45) : hsl(220,25,65,0.4);
    ctx.lineWidth = i%3===0?2:1; ctx.lineCap='round'; ctx.stroke();
  }

  const sep0 = tickOuter - 10;
  ctx.beginPath(); ctx.arc(CX,CY,sep0,0,Math.PI*2);
  ctx.strokeStyle = isAM ? hsl(0,0,0,0.07) : hsl(0,0,100,0.07);
  ctx.lineWidth=1; ctx.stroke();

  drawSpokes(centerR, sep0, isAM);

  drawRingActivities(activities[1], r1, w1, isAM?1:0.4, h, 1);
  const sep1 = r1 - w1/2 - 1;
  ctx.beginPath(); ctx.arc(CX,CY,sep1,0,Math.PI*2);
  ctx.strokeStyle = isAM ? hsl(0,0,0,0.07) : hsl(0,0,100,0.07);
  ctx.lineWidth=0.5; ctx.stroke();

  drawRingActivities(activities[2], r2, w2, isAM?0.4:1, h, 2);
  const sep2 = r2 - w2/2 - 1;
  ctx.beginPath(); ctx.arc(CX,CY,sep2,0,Math.PI*2);
  ctx.strokeStyle = isAM ? hsl(0,0,0,0.07) : hsl(0,0,100,0.07);
  ctx.lineWidth=0.5; ctx.stroke();

  const cr = Math.max(centerR, 8);
  ctx.beginPath(); ctx.arc(CX,CY,cr,0,Math.PI*2);
  ctx.fillStyle = isAM ? hsl(48,60,93) : hsl(228,30,7); ctx.fill();
  ctx.strokeStyle = isAM ? hsl(45,40,68,0.4) : hsl(228,30,38,0.4);
  ctx.lineWidth=1; ctx.stroke();

  if (cr > 8) {
    ctx.font = `${Math.min(cr*1.2,20)}px serif`;
    ctx.textAlign='center'; ctx.textBaseline='middle';
    ctx.fillText(isAM?'☀️':'🌙', CX, CY);
  }

  const minLen  = NUM_R - 10;
  const hourLen = NUM_R - 10 - (W_THICK + W_THIN + GAP) * 0.55;
  const h12 = (h%12)+m/60;
  const hourA = hourToAngle(h12);
  const minA  = (m/60)*Math.PI*2 - Math.PI/2;
  const handClr = isAM ? hsl(30,55,22) : hsl(220,55,90);
  const minClr  = isAM ? hsl(30,45,38) : hsl(220,40,75);

  ctx.beginPath(); ctx.moveTo(CX,CY);
  ctx.lineTo(CX+hourLen*Math.cos(hourA), CY+hourLen*Math.sin(hourA));
  ctx.strokeStyle=handClr; ctx.lineWidth=4; ctx.lineCap='round'; ctx.stroke();

  ctx.beginPath(); ctx.moveTo(CX,CY);
  ctx.lineTo(CX+minLen*Math.cos(minA), CY+minLen*Math.sin(minA));
  ctx.strokeStyle=minClr; ctx.lineWidth=2.5; ctx.lineCap='round'; ctx.stroke();

  ctx.beginPath(); ctx.arc(CX,CY,3.5,0,Math.PI*2);
  ctx.fillStyle=handClr; ctx.fill();
}

renderLists();
draw();
</script>
</body>
</html>
```

---

## 11. Open Questions

| Topic | Status | Options |
|---|---|---|
| Auth scope | ✅ JWT + refresh token | — |
| Multiple parents per child | ❌ out of MVP scope | many-to-many `user_children` in v2 |
| Frontend framework | ❓ TBD | plain JS / HTMX / Vue |
| Hosting | ❓ TBD | VPS / Docker |
| Upload storage | ❓ TBD | local disk / S3 |
| Push notifications | ❌ out of MVP scope | `ActivityStartingSoon` event in v2 |
| Seasonal schedule templates | ❌ out of MVP scope | `ScheduleTemplate` + `ActivateSchedule` in v2 |
| Overlapping events | ⚠️ warning only (not a hard block) | validation in service layer |

---

## 12. Next Steps

1. SQL migrations (001–007)
2. Domain models in Go (`internal/domain/`)
3. Repository interfaces
4. Repository implementations with sqlx
5. Application services
6. HTTP handlers + routing
7. Auth middleware (JWT)
8. Frontend — clock view (canvas)
9. Frontend — configurator
10. Integration tests for critical paths