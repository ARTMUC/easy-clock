package i18n

const (
	// auth flow
	MsgFillAllFields           = "fill_all_fields"
	MsgNotActive               = "not_active"
	MsgInvalidCredentials      = "invalid_credentials"
	MsgServerError             = "server_error"
	MsgEmailTaken              = "email_taken"
	MsgVerificationEmailFailed = "verification_email_failed"
	MsgInvalidToken            = "invalid_token"
	MsgAccountActivated        = "account_activated"

	// login page
	MsgLoginTitle       = "login_title"
	MsgLoginSubtitle    = "login_subtitle"
	MsgLabelEmail       = "label_email"
	MsgLabelPassword    = "label_password"
	MsgBtnLogin         = "btn_login"
	MsgNoAccount        = "no_account"
	MsgSignUp           = "sign_up"

	// register page
	MsgRegisterTitle    = "register_title"
	MsgRegisterSubtitle = "register_subtitle"
	MsgLabelName        = "label_name"
	MsgBtnCreateAccount = "btn_create_account"
	MsgHaveAccount      = "have_account"

	// verify page
	MsgVerifyTitle      = "verify_title"

	// check email page
	MsgCheckEmailTitle  = "check_email_title"
	MsgCheckEmailBody   = "check_email_body"

	// dashboard page
	MsgChildren         = "children"
	MsgNoChildren       = "no_children"
	MsgClock            = "clock"
	MsgConfigure        = "configure"
	MsgAddChild         = "add_child"
	MsgLabelTimezone    = "label_timezone"
	MsgPlaceholderName  = "placeholder_name"
	MsgPlaceholderTZ    = "placeholder_tz"

	// child config page
	MsgDeleteChild      = "delete_child"
	MsgDeleteChildConfirm = "delete_child_confirm"
	MsgProfiles         = "profiles"
	MsgNoProfiles       = "no_profiles"
	MsgDefaultProfile   = "default_profile"
	MsgSetDefault       = "set_default"
	MsgAddProfile       = "add_profile"
	MsgProfileName      = "profile_name"
	MsgWeeklySchedule   = "weekly_schedule"
	MsgSave             = "save"
	MsgBackDashboard    = "back_dashboard"
	MsgDeleteConfirm    = "delete_confirm"

	// schedule day names
	MsgDaySun = "day_sun"
	MsgDayMon = "day_mon"
	MsgDayTue = "day_tue"
	MsgDayWed = "day_wed"
	MsgDayThu = "day_thu"
	MsgDayFri = "day_fri"
	MsgDaySat = "day_sat"
	MsgDayNone = "day_none"

	// profile config page
	MsgActivities       = "activities"
	MsgBackToChild      = "back_to_child"
	MsgRing1            = "ring1"
	MsgRing2            = "ring2"
	MsgNoneYet          = "none_yet"
	MsgAddActivity      = "add_activity"
	MsgLabelEmoji       = "label_emoji"
	MsgLabelLabel       = "label_label"
	MsgLabelImagePath   = "label_image_path"
	MsgLabelFromHour    = "label_from_hour"
	MsgLabelToHour      = "label_to_hour"
	MsgLabelRing        = "label_ring"
	MsgLabelSortOrder   = "label_sort_order"

	// clock page
	MsgNoSchedule       = "no_schedule"
)

var messages = map[string]map[Lang]string{
	MsgFillAllFields: {
		PL: "Wypełnij wszystkie pola.",
		EN: "Please fill in all fields.",
	},
	MsgNotActive: {
		PL: "Konto nie zostało aktywowane. Sprawdź skrzynkę e-mail.",
		EN: "Account not activated. Please check your email.",
	},
	MsgInvalidCredentials: {
		PL: "Nieprawidłowy e-mail lub hasło.",
		EN: "Invalid email or password.",
	},
	MsgServerError: {
		PL: "Błąd serwera. Spróbuj ponownie.",
		EN: "Server error. Please try again.",
	},
	MsgEmailTaken: {
		PL: "Ten adres e-mail jest już zajęty.",
		EN: "This email is already taken.",
	},
	MsgVerificationEmailFailed: {
		PL: "Nie udało się wysłać e-maila weryfikacyjnego. Sprawdź adres lub spróbuj później.",
		EN: "Failed to send verification email. Check your address or try again.",
	},
	MsgInvalidToken: {
		PL: "Link weryfikacyjny jest nieprawidłowy lub wygasł.",
		EN: "The verification link is invalid or has expired.",
	},
	MsgAccountActivated: {
		PL: "Konto zostało aktywowane. Możesz się teraz zalogować.",
		EN: "Your account has been activated. You can now log in.",
	},
	MsgLoginTitle: {PL: "Zaloguj się — KidClock", EN: "Log in — KidClock"},
	MsgLoginSubtitle: {PL: "Zaloguj się do konta", EN: "Sign in to your account"},
	MsgLabelEmail:    {PL: "E-mail", EN: "Email"},
	MsgLabelPassword: {PL: "Hasło", EN: "Password"},
	MsgBtnLogin:      {PL: "Zaloguj się", EN: "Log in"},
	MsgNoAccount:     {PL: "Nie masz konta?", EN: "Don't have an account?"},
	MsgSignUp:        {PL: "Zarejestruj się", EN: "Sign up"},

	MsgRegisterTitle:    {PL: "Rejestracja — KidClock", EN: "Register — KidClock"},
	MsgRegisterSubtitle: {PL: "Utwórz konto", EN: "Create your account"},
	MsgLabelName:        {PL: "Imię", EN: "Name"},
	MsgBtnCreateAccount: {PL: "Utwórz konto", EN: "Create account"},
	MsgHaveAccount:      {PL: "Masz już konto?", EN: "Already have an account?"},

	MsgVerifyTitle: {PL: "Weryfikacja e-mail — KidClock", EN: "Verify email — KidClock"},

	MsgCheckEmailTitle: {PL: "Sprawdź pocztę — KidClock", EN: "Check your email — KidClock"},
	MsgCheckEmailBody: {
		PL: "Wysłaliśmy link aktywacyjny na adres",
		EN: "We sent a verification link to",
	},

	MsgChildren:        {PL: "Dzieci", EN: "Children"},
	MsgNoChildren:      {PL: "Brak dzieci. Dodaj pierwsze poniżej.", EN: "No children yet. Add one below."},
	MsgClock:           {PL: "Zegar ↗", EN: "Clock ↗"},
	MsgConfigure:       {PL: "Konfiguruj", EN: "Configure"},
	MsgAddChild:        {PL: "Dodaj dziecko", EN: "Add child"},
	MsgLabelTimezone:   {PL: "Strefa czasowa", EN: "Timezone"},
	MsgPlaceholderName: {PL: "np. Ania", EN: "e.g. Anna"},
	MsgPlaceholderTZ:   {PL: "np. Europe/Warsaw", EN: "e.g. Europe/London"},

	MsgDeleteChild:        {PL: "Usuń dziecko", EN: "Delete child"},
	MsgDeleteChildConfirm: {PL: "Usunąć dziecko?", EN: "Delete child?"},
	MsgProfiles:           {PL: "Profile", EN: "Profiles"},
	MsgNoProfiles:         {PL: "Brak profili.", EN: "No profiles yet."},
	MsgDefaultProfile:     {PL: "(domyślny)", EN: "(default)"},
	MsgSetDefault:         {PL: "Ustaw domyślny", EN: "Set default"},
	MsgAddProfile:         {PL: "Dodaj profil", EN: "Add profile"},
	MsgProfileName:        {PL: "Nazwa profilu", EN: "Profile name"},
	MsgWeeklySchedule:     {PL: "Harmonogram tygodniowy", EN: "Weekly schedule"},
	MsgSave:               {PL: "Zapisz", EN: "Save"},
	MsgBackDashboard:      {PL: "← Pulpit", EN: "← Dashboard"},
	MsgDeleteConfirm:      {PL: "Usunąć?", EN: "Delete?"},

	MsgDaySun:  {PL: "Nie", EN: "Sun"},
	MsgDayMon:  {PL: "Pon", EN: "Mon"},
	MsgDayTue:  {PL: "Wt", EN: "Tue"},
	MsgDayWed:  {PL: "Śr", EN: "Wed"},
	MsgDayThu:  {PL: "Czw", EN: "Thu"},
	MsgDayFri:  {PL: "Pt", EN: "Fri"},
	MsgDaySat:  {PL: "Sob", EN: "Sat"},
	MsgDayNone: {PL: "— brak —", EN: "— none —"},

	MsgActivities:     {PL: "aktywności", EN: "activities"},
	MsgBackToChild:    {PL: "← Wstecz", EN: "← Back"},
	MsgRing1:          {PL: "Pierścień 1 (rano)", EN: "Ring 1 (AM)"},
	MsgRing2:          {PL: "Pierścień 2 (popołudnie)", EN: "Ring 2 (PM)"},
	MsgNoneYet:        {PL: "Brak.", EN: "None yet."},
	MsgAddActivity:    {PL: "Dodaj aktywność", EN: "Add activity"},
	MsgLabelEmoji:     {PL: "Emoji", EN: "Emoji"},
	MsgLabelLabel:     {PL: "Nazwa", EN: "Label"},
	MsgLabelImagePath: {PL: "Ścieżka do obrazka", EN: "Image path"},
	MsgLabelFromHour:  {PL: "Godzina od (0–23)", EN: "From hour (0–23)"},
	MsgLabelToHour:    {PL: "Godzina do (1–24)", EN: "To hour (1–24)"},
	MsgLabelRing:      {PL: "Pierścień", EN: "Ring"},
	MsgLabelSortOrder: {PL: "Kolejność", EN: "Sort order"},

	MsgNoSchedule: {PL: "Brak harmonogramu.", EN: "No schedule for now."},
}

// Msg returns the translated string for key in lang, falling back to EN.
func Msg(key string, l Lang) string {
	if m, ok := messages[key]; ok {
		if s, ok := m[l]; ok && s != "" {
			return s
		}
		return m[EN]
	}
	return key
}

// DayName returns the short weekday name for day 0–6 (0=Sun).
func DayName(day int, l Lang) string {
	keys := [7]string{MsgDaySun, MsgDayMon, MsgDayTue, MsgDayWed, MsgDayThu, MsgDayFri, MsgDaySat}
	if day < 0 || day > 6 {
		return ""
	}
	return Msg(keys[day], l)
}
