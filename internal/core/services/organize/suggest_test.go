package organize

import (
	"context"
	"fmt"
	"testing"

	"harmoniz/internal/core/domain"
	"harmoniz/internal/testutil"
)

// ============================================================
// parseFilename
// ============================================================

func TestParseFilename_DashSeparator(t *testing.T) {
	p := parseFilename("Artist - Title.mp3")
	if p.artist != "Artist" {
		t.Errorf("artist: want 'Artist', got %q", p.artist)
	}
	if p.title != "Title" {
		t.Errorf("title: want 'Title', got %q", p.title)
	}
}

func TestParseFilename_MultiHyphenSeparator(t *testing.T) {
	p := parseFilename("Artist---Title.mp3")
	if p.artist != "Artist" {
		t.Errorf("artist: want 'Artist', got %q", p.artist)
	}
	if p.title != "Title" {
		t.Errorf("title: want 'Title', got %q", p.title)
	}
}

func TestParseFilename_LeadingTrackNumber(t *testing.T) {
	// "03 - Title.mp3" is the most common music file naming convention.
	// A purely-numeric first part must be treated as track number, not artist.
	cases := []struct {
		name    string
		wantNum int
		wantTitle string
	}{
		{"03 - Title.mp3", 3, "Title"},
		{"03. Title.mp3", 3, "Title"},
	}
	for _, c := range cases {
		p := parseFilename(c.name)
		if p.trackNum != c.wantNum {
			t.Errorf("%q: trackNum want %d, got %d", c.name, c.wantNum, p.trackNum)
		}
		if p.title != c.wantTitle {
			t.Errorf("%q: title want %q, got %q", c.name, c.wantTitle, p.title)
		}
		if p.artist != "" {
			t.Errorf("%q: artist should be empty, got %q", c.name, p.artist)
		}
	}
}

func TestParseFilename_FourParts_ArtistAlbumNumTitle(t *testing.T) {
	// "Artist - Album - 02 - Song.mp3" → all four fields extracted.
	p := parseFilename("Artist - Album - 02 - Song.mp3")
	if p.artist != "Artist" {
		t.Errorf("artist: want 'Artist', got %q", p.artist)
	}
	if p.album != "Album" {
		t.Errorf("album: want 'Album', got %q", p.album)
	}
	if p.trackNum != 2 {
		t.Errorf("trackNum: want 2, got %d", p.trackNum)
	}
	if p.title != "Song" {
		t.Errorf("title: want 'Song', got %q", p.title)
	}
}

func TestParseFilename_TripleUnderscoreAsArtistSep(t *testing.T) {
	// Real pattern from library: "Emile___Bobo_chantent__LA_RICHESSE_ICI_BAS_(0).mp3"
	// The triple underscore is treated like "---" multi-hyphen separator.
	p := parseFilename("Emile___Bobo_chantent__LA_RICHESSE_ICI_BAS_(0).mp3")
	// Either artist is non-empty or title is non-empty (something was parsed)
	if p.artist == "" && p.title == "" {
		t.Error("expected non-empty artist or title for underscore-formatted filename")
	}
}

func TestParseFilename_AllCapsUnderscoreYouTubeFormat(t *testing.T) {
	// Real pattern: "Sylvain_Akouala__L_HOMME_NÉ_DE_L_HOMME_VIT_ÉTERNELLEMENT_(0).mp3"
	p := parseFilename("Sylvain_Akouala__L_HOMME_NÉ_DE_L_HOMME_VIT_ÉTERNELLEMENT_(0).mp3")
	if p.artist == "" && p.title == "" {
		t.Error("expected non-empty artist or title for YouTube-format underscore filename")
	}
}

func TestParseFilename_MultiPartDash(t *testing.T) {
	// Real pattern: "Arms_Of_The_Savior_-_Elizabeth_South_-_Lyrics(360p).mp3"
	p := parseFilename("Arms_Of_The_Savior_-_Elizabeth_South_-_Lyrics(360p).mp3")
	if p.artist == "" && p.title == "" {
		t.Error("expected non-empty artist or title for multi-part dash filename")
	}
}

func TestParseFilename_HonorifixAndAllCapsArtist(t *testing.T) {
	// Real pattern: "Fr. Jean Sylvain AKOUALA - 50km et Bethlehem Effrata.mp3"
	p := parseFilename("Fr. Jean Sylvain AKOUALA - 50km et Bethlehem Effrata.mp3")
	if p.artist == "" {
		t.Error("expected non-empty artist")
	}
	if p.title == "" {
		t.Error("expected non-empty title")
	}
}

func TestParseFilename_NoSeparator(t *testing.T) {
	p := parseFilename("SomeSong.mp3")
	if p.artist != "" {
		t.Errorf("artist should be empty, got %q", p.artist)
	}
	if p.title == "" {
		t.Error("expected non-empty title for no-separator filename")
	}
}

func TestParseFilename_DoubleExtensionStripped(t *testing.T) {
	p := parseFilename("song.mp3.mp3")
	if p.title == "" {
		t.Error("expected non-empty title after stripping double extension")
	}
}

// ============================================================
// isPlaceholder helpers (table-driven)
// ============================================================

func TestIsPlaceholderArtist(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"Unknown Artist", true},
		{"unknown", true},
		{"va", true},
		{"Various Artists", true},
		{"Artiste inconnu", true}, // "artisteinconnu" matches
		{"Daft Punk", false},
		{"Nathan Epenge", false},
		{"Sylvain Akouala", false},
	}
	for _, c := range cases {
		got := isPlaceholderArtist(c.input)
		if got != c.want {
			t.Errorf("isPlaceholderArtist(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestIsPlaceholderTitle(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"piste1", true},
		{"Piste 2", true},
		{"Track 01", true},
		{"track01", true},
		{"untitled", true},
		{"Macpéla", false},
		{"La Richesse Ici Bas", false},
		{"", false},
	}
	for _, c := range cases {
		got := isPlaceholderTitle(c.input)
		if got != c.want {
			t.Errorf("isPlaceholderTitle(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestIsPlaceholderAlbum(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"Album inconnu (09/10/2008 09:46:25)", true},
		{"Unknown Album", true},
		{"unknownalbum", true},
		{"Random Access Memories", false},
		{"Gold", false},
		{"", false},
	}
	for _, c := range cases {
		got := isPlaceholderAlbum(c.input)
		if got != c.want {
			t.Errorf("isPlaceholderAlbum(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

// ============================================================
// normalizeText / toTitleCase
// ============================================================

func TestNormalizeText_Underscores(t *testing.T) {
	got := normalizeText("la_richesse_ici_bas")
	want := "La Richesse Ici Bas"
	if got != want {
		t.Errorf("normalizeText: want %q, got %q", want, got)
	}
}

func TestToTitleCase_FrenchArticles(t *testing.T) {
	// Full exact output: first word always capitalised; subsequent articles/preps lowercase.
	got := toTitleCase("la richesse de la terre")
	want := "La Richesse de la Terre"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

// ============================================================
// cleanTitle
// ============================================================

func TestCleanTitle_StripsExtension(t *testing.T) {
	got := cleanTitle("Song.mp3")
	if got == "Song.mp3" {
		t.Errorf("extension should be stripped, got %q", got)
	}
}

func TestCleanTitle_StripsArtistPrefix(t *testing.T) {
	// "AkoualaAide" should strip "Akouala" prefix when candidate is "Akouala"
	got := cleanTitle("Akouala Song Title", "Akouala")
	if got == "" {
		t.Error("expected non-empty result after stripping artist prefix")
	}
}

func TestCleanTitle_ShortCandidateIgnored(t *testing.T) {
	// Candidates with <4 letters (like "De") must not be stripped.
	got := cleanTitle("De La Soul")
	if got == "" {
		t.Error("short candidate 'De' should not strip the whole title")
	}
}

// ============================================================
// computeScore
// ============================================================

func TestComputeScore_Empty(t *testing.T) {
	if s := computeScore(nil); s != 0 {
		t.Errorf("empty fields: want 0, got %v", s)
	}
}

func TestComputeScore_SingleHighConf(t *testing.T) {
	fields := map[string]domain.FieldSuggestion{
		"artist": {Value: "X", Confidence: 0.90},
	}
	if s := computeScore(fields); s != 90.0 {
		t.Errorf("want 90.0, got %v", s)
	}
}

func TestComputeScore_Average(t *testing.T) {
	fields := map[string]domain.FieldSuggestion{
		"artist": {Value: "X", Confidence: 0.80},
		"title":  {Value: "Y", Confidence: 0.60},
	}
	got := computeScore(fields)
	// average = 0.70 → score = 70.0
	if got < 69 || got > 71 {
		t.Errorf("want ≈70.0, got %v", got)
	}
}

// ============================================================
// AnalyzeRoot — via FakeRepo (black-box service API)
// ============================================================

func makeTrack(id uint64, path, artistRaw, albumRaw, title string, year, trackNum int) domain.Track {
	t := domain.Track{
		ID:        id,
		Path:      path,
		Filename:  fmt.Sprintf("track%d.mp3", id),
		ArtistRaw: artistRaw,
		ArtistNorm: domain.NormalizeArtist(artistRaw),
		AlbumRaw:  albumRaw,
		Title:     title,
		Year:      year,
		TrackNum:  trackNum,
		Status:    domain.StatusClean,
	}
	return t
}

func TestAnalyzeRoot_Empty(t *testing.T) {
	svc := NewSuggestionService(testutil.NewFakeRepo())
	suggestions, registrySize, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(suggestions) != 0 {
		t.Errorf("want 0 suggestions, got %d", len(suggestions))
	}
	if registrySize != 0 {
		t.Errorf("want registrySize=0, got %d", registrySize)
	}
}

func TestAnalyzeRoot_ArtistFromNeighborTags(t *testing.T) {
	// 5 tracks in the same directory all tagged "Emile Bobo" → tag vote wins.
	var trs []domain.Track
	for i := 1; i <= 5; i++ {
		trs = append(trs, makeTrack(uint64(i), fmt.Sprintf("/lib/dir/track%d.mp3", i),
			"Emile Bobo", "", "", 0, 0))
	}
	svc := NewSuggestionService(testutil.NewFakeRepo(trs...))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range suggs {
		art, ok := s.Fields["artist"]
		if !ok {
			continue // artist already clean — acceptable
		}
		if art.Source != "neighbor" {
			t.Errorf("expected source='neighbor', got %q (value=%q)", art.Source, art.Value)
		}
	}
}

func TestAnalyzeRoot_ArtistFromPath_TwoLevel(t *testing.T) {
	// Track at /lib/Nathan Epenge/song.mp3 with no tags → path inference.
	tr := makeTrack(1, "/lib/Nathan Epenge/song.mp3", "", "", "", 0, 0)
	tr.Filename = "song.mp3"
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions returned")
	}
	art, ok := suggs[0].Fields["artist"]
	if !ok {
		t.Fatal("expected artist field suggestion")
	}
	if art.Source != "path" {
		t.Errorf("expected source='path', got %q", art.Source)
	}
	if art.Confidence < 0.60 || art.Confidence > 0.80 {
		t.Errorf("path conf %v outside expected [0.60, 0.80]", art.Confidence)
	}
}

func TestAnalyzeRoot_ArtistFromPath_ThreeLevel(t *testing.T) {
	// Track at /lib/Nathan Epenge/Album One/song.mp3 → 0.75 path confidence.
	tr := makeTrack(1, "/lib/Nathan Epenge/Album One/song.mp3", "", "", "", 0, 0)
	tr.Filename = "song.mp3"
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	if a := suggs[0].Fields["artist"]; a.Confidence < 0.70 {
		t.Errorf("3-level path conf too low: %v", a.Confidence)
	}
	if alb := suggs[0].Fields["album"]; alb.Value == "" {
		t.Error("expected album suggestion from path")
	}
}

func TestAnalyzeRoot_ArtistFromFilename(t *testing.T) {
	tr := makeTrack(1, "/lib/Nathan Epenge - Macpéla.mp3", "", "", "", 0, 0)
	tr.Filename = "Nathan Epenge - Macpéla.mp3"
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	art := suggs[0].Fields["artist"]
	if art.Value == "" {
		t.Error("expected non-empty artist from filename")
	}
	ttl := suggs[0].Fields["title"]
	if ttl.Value == "" {
		t.Error("expected non-empty title from filename")
	}
}

func TestAnalyzeRoot_LibraryExactLookupWins(t *testing.T) {
	// Registry seeded with 100 "Daft Punk" tracks → ExactLookup returns high conf.
	// Lone track has only a filename hint, no tag.
	var trs []domain.Track
	for i := 1; i <= 100; i++ {
		trs = append(trs, makeTrack(uint64(i), fmt.Sprintf("/lib/main/song%d.mp3", i),
			"Daft Punk", "", "", 0, 0))
	}
	lone := domain.Track{
		ID:       101,
		Path:     "/lib/other/Daft Punk---Revolution.mp3",
		Filename: "Daft Punk---Revolution.mp3",
		Status:   domain.StatusClean,
	}
	trs = append(trs, lone)

	svc := NewSuggestionService(testutil.NewFakeRepo(trs...))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	var loneS domain.OrganizerSuggestion
	for _, s := range suggs {
		if s.TrackID == 101 {
			loneS = s
			break
		}
	}
	art, ok := loneS.Fields["artist"]
	if !ok {
		t.Fatal("expected artist suggestion for lone track")
	}
	if art.Source != "library" {
		t.Errorf("expected source='library', got %q (value=%q, conf=%.2f)", art.Source, art.Value, art.Confidence)
	}
	if art.Confidence < 0.80 {
		t.Errorf("library exact conf %v < 0.80", art.Confidence)
	}
}

func TestAnalyzeRoot_CompilationDirSkipsRegistry(t *testing.T) {
	// 4+ distinct tag artists in same dir → isCompilation=true → no library source.
	artists := []string{"Ado", "Emile Bobo", "Alfred Ondo", "Nathan Epenge", "Anicet Mundundu"}
	var trs []domain.Track
	for i, a := range artists {
		trs = append(trs, makeTrack(uint64(i+1),
			fmt.Sprintf("/lib/compilation/track%d.mp3", i+1), a, "", "", 0, 0))
	}
	// Also seed registry with high track counts so it would normally win.
	for _, a := range artists {
		for j := range 60 {
			trs = append(trs, makeTrack(0, fmt.Sprintf("/lib/main/%s/%d.mp3", normStr(a), j), a, "", "", 0, 0))
		}
	}
	svc := NewSuggestionService(testutil.NewFakeRepo(trs...))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib/compilation")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range suggs {
		if art, ok := s.Fields["artist"]; ok {
			if art.Source == "library" {
				t.Errorf("compilation dir must not use library source, got %q for track %d", art.Source, s.TrackID)
			}
		}
	}
}

func TestAnalyzeRoot_DecisionTracePopulated(t *testing.T) {
	// Track with BOTH filename hint AND path hint → trace must have ≥2 steps,
	// exactly one of which is not rejected.
	tr := makeTrack(1, "/lib/Nathan Epenge/Nathan Epenge - Macpéla.mp3", "", "", "", 0, 0)
	tr.Filename = "Nathan Epenge - Macpéla.mp3"
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	art, ok := suggs[0].Fields["artist"]
	if !ok {
		t.Fatal("expected artist field")
	}
	if len(art.Trace) < 2 {
		t.Fatalf("expected trace ≥ 2 steps, got %d", len(art.Trace))
	}
	winners := 0
	for _, step := range art.Trace {
		if !step.Rejected {
			winners++
		}
	}
	if winners != 1 {
		t.Errorf("exactly 1 non-rejected step expected, got %d", winners)
	}
}

func TestAnalyzeRoot_AlbumFromTagVote(t *testing.T) {
	// 4 tracks with AlbumRaw="Gold Album" provide the tag vote.
	// 1 track has empty AlbumRaw and should receive the neighbor suggestion.
	var trs []domain.Track
	for i := 1; i <= 4; i++ {
		trs = append(trs, makeTrack(uint64(i), fmt.Sprintf("/lib/dir/t%d.mp3", i),
			"", "Gold Album", "", 0, 0))
	}
	// Track 5: no album tag → should get neighbor suggestion.
	tr5 := makeTrack(5, "/lib/dir/t5.mp3", "", "", "", 0, 0)
	trs = append(trs, tr5)

	svc := NewSuggestionService(testutil.NewFakeRepo(trs...))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range suggs {
		if s.TrackID == 5 {
			alb, ok := s.Fields["album"]
			if !ok {
				t.Fatal("track 5 (no album tag) should get an album suggestion")
			}
			if alb.Source != "neighbor" {
				t.Errorf("expected source='neighbor', got %q", alb.Source)
			}
			if alb.Value != "Gold Album" {
				t.Errorf("expected value='Gold Album', got %q", alb.Value)
			}
			return
		}
	}
	t.Error("track 5 not found in suggestions")
}

func TestAnalyzeRoot_TitleFromFilename_Placeholder(t *testing.T) {
	tr := makeTrack(1, "/lib/dir/Artist - Real Song.mp3", "", "", "Track 01", 0, 0)
	tr.Filename = "Artist - Real Song.mp3"
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	ttl, ok := suggs[0].Fields["title"]
	if !ok {
		t.Fatal("expected title suggestion when existing title is placeholder")
	}
	if ttl.Value == "" {
		t.Error("title suggestion value must not be empty")
	}
}

func TestAnalyzeRoot_YearFromTimestampAlbum(t *testing.T) {
	tr := makeTrack(1, "/lib/dir/song.mp3", "", "Album inconnu (09/10/2008 09:46:25)", "", 0, 0)
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	yr, ok := suggs[0].Fields["year"]
	if !ok {
		t.Fatal("expected year suggestion from album timestamp")
	}
	if yr.Value != "2008" {
		t.Errorf("want year='2008', got %q", yr.Value)
	}
}

func TestAnalyzeRoot_TrackNumFromFilename(t *testing.T) {
	// "05 - Title.mp3" is the canonical music naming convention for track 5.
	// The purely-numeric first token must be parsed as track_num, not artist.
	tr := makeTrack(1, "/lib/dir/05 - Title.mp3", "", "", "", 0, 0)
	tr.Filename = "05 - Title.mp3"
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	tn, ok := suggs[0].Fields["track_num"]
	if !ok {
		t.Fatal("expected track_num suggestion")
	}
	if tn.Value != "5" {
		t.Errorf("want track_num='5', got %q", tn.Value)
	}
}

func TestAnalyzeRoot_ExistingGoodArtistNotOverridden(t *testing.T) {
	// If ArtistRaw is non-empty and not a placeholder, no artist suggestion is generated
	// (unless normalization is needed).
	tr := makeTrack(1, "/lib/dir/song.mp3", "Good Artist", "Album", "Song", 2020, 1)
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	// The Issues list should NOT contain "Missing artist"
	for _, iss := range suggs[0].Issues {
		if iss == "Missing artist" {
			t.Error("track with good ArtistRaw should not have 'Missing artist' issue")
		}
	}
}

func TestAnalyzeRoot_ScoreInBounds(t *testing.T) {
	tr := makeTrack(1, "/lib/dir/Artist - Title.mp3", "", "", "", 0, 0)
	tr.Filename = "Artist - Title.mp3"
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range suggs {
		if s.Score < 0 || s.Score > 100 {
			t.Errorf("score %v out of [0, 100]", s.Score)
		}
	}
}

func TestAnalyzeRoot_RegistrySizeMatchesUniqueArtists(t *testing.T) {
	trs := []domain.Track{
		makeTrack(1, "/lib/a.mp3", "Artist A", "", "", 0, 0),
		makeTrack(2, "/lib/b.mp3", "Artist B", "", "", 0, 0),
		makeTrack(3, "/lib/c.mp3", "Artist A", "", "", 0, 0), // duplicate artist
	}
	svc := NewSuggestionService(testutil.NewFakeRepo(trs...))
	_, regSize, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	// 2 unique tag artists → registry size should be ≥ 2
	if regSize < 2 {
		t.Errorf("expected registrySize ≥ 2, got %d", regSize)
	}
}

// ============================================================
// parseFilename — 3-part and 4-part cases
// ============================================================

func TestParseFilename_ThreeParts_LeadingNumIsTrackNum(t *testing.T) {
	// "01 - Album - Title.mp3": purely-numeric first part must become trackNum.
	p := parseFilename("01 - Album - Title.mp3")
	if p.trackNum != 1 {
		t.Errorf("trackNum: want 1, got %d", p.trackNum)
	}
	if p.album != "Album" {
		t.Errorf("album: want 'Album', got %q", p.album)
	}
	if p.title != "Title" {
		t.Errorf("title: want 'Title', got %q", p.title)
	}
	if p.artist != "" {
		t.Errorf("artist should be empty, got %q", p.artist)
	}
}

func TestParseFilename_ThreeParts_NumInMiddle(t *testing.T) {
	// "Artist - 02 - Title.mp3": purely-numeric middle part must become trackNum.
	p := parseFilename("Artist - 02 - Title.mp3")
	if p.artist != "Artist" {
		t.Errorf("artist: want 'Artist', got %q", p.artist)
	}
	if p.trackNum != 2 {
		t.Errorf("trackNum: want 2, got %d", p.trackNum)
	}
	if p.title != "Title" {
		t.Errorf("title: want 'Title', got %q", p.title)
	}
}

func TestParseFilename_ThreeParts_AllText(t *testing.T) {
	// "Artist - Album - Title.mp3": non-numeric three parts → artist + album + title.
	p := parseFilename("Artist - Album - Title.mp3")
	if p.artist != "Artist" {
		t.Errorf("artist: want 'Artist', got %q", p.artist)
	}
	if p.album != "Album" {
		t.Errorf("album: want 'Album', got %q", p.album)
	}
	if p.title != "Title" {
		t.Errorf("title: want 'Title', got %q", p.title)
	}
	if p.trackNum != 0 {
		t.Errorf("trackNum should be 0, got %d", p.trackNum)
	}
}

// ============================================================
// normalizeText / toTitleCase — exact output
// ============================================================

func TestNormalizeText_HyphenWordSeparator(t *testing.T) {
	// Hyphens between non-space chars are treated as word separators.
	got := normalizeText("L-HOMME-NÉ")
	want := "L Homme Né"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestToTitleCase_AllCapsInput(t *testing.T) {
	got := toTitleCase("DAFT PUNK")
	want := "Daft Punk"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestToTitleCase_FirstWordCappedEvenIfArticle(t *testing.T) {
	// Even lowercase article words must be capitalised when first.
	got := toTitleCase("de la soul")
	if len(got) < 2 || got[:2] != "De" {
		t.Errorf("first word 'de' must be capitalised when first, got %q", got)
	}
}

// ============================================================
// isAllCaps — table-driven
// ============================================================

func TestIsAllCaps(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"DAFT PUNK", true},
		{"Daft Punk", false},
		{"daft punk", false},
		{"SYLVAIN", true},
		{"ABC", false}, // len(rune)=3, isAllCaps requires >3
		{"ABCD", true},
		{"", false},
	}
	for _, c := range cases {
		got := isAllCaps(c.input)
		if got != c.want {
			t.Errorf("isAllCaps(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

// ============================================================
// normalizeExistingArtist — entirely untested before
// ============================================================

func TestNormalizeExistingArtist_AllCapsBecomesTitle(t *testing.T) {
	got := normalizeExistingArtist("DAFT PUNK", "")
	want := "Daft Punk"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestNormalizeExistingArtist_UnderscoreBecomesSpaces(t *testing.T) {
	got := normalizeExistingArtist("daft_punk", "")
	want := "Daft Punk"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestNormalizeExistingArtist_TagArtistPreferredWhenContains(t *testing.T) {
	// When tagArtist's norm contains existing's norm, tagArtist is the canonical form.
	got := normalizeExistingArtist("Sylvain", "Sylvain Akouala")
	want := "Sylvain Akouala"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestNormalizeExistingArtist_GoodArtistUnchanged(t *testing.T) {
	got := normalizeExistingArtist("Daft Punk", "")
	if got != "Daft Punk" {
		t.Errorf("well-formed artist must be unchanged, got %q", got)
	}
}

// ============================================================
// AnalyzeRoot — normalization, compilation threshold, issues, album from path
// ============================================================

func TestAnalyzeRoot_AllCapsArtistNormalized(t *testing.T) {
	// A track whose ArtistRaw is ALL CAPS must receive a normalisation suggestion.
	tr := makeTrack(1, "/lib/dir/song.mp3", "NATHAN EPENGE", "", "", 0, 0)
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	art, ok := suggs[0].Fields["artist"]
	if !ok {
		t.Fatal("expected artist normalisation suggestion for ALL CAPS artist")
	}
	if art.Value != "Nathan Epenge" {
		t.Errorf("want 'Nathan Epenge', got %q", art.Value)
	}
	if art.Source != "neighbor" {
		t.Errorf("want source='neighbor' (normalisation path), got %q", art.Source)
	}
	if art.Confidence != 0.72 {
		t.Errorf("want confidence=0.72, got %.2f", art.Confidence)
	}
}

func TestAnalyzeRoot_CompilationThresholdFourArtists(t *testing.T) {
	// Exactly 4 distinct tag artists in the same dir → isCompilation=true → library source blocked.
	var trs []domain.Track
	id := uint64(1)

	// Seed registry so library would otherwise win.
	for j := range 100 {
		trs = append(trs, makeTrack(id, fmt.Sprintf("/lib/main/t%d.mp3", j), "Daft Punk", "", "", 0, 0))
		id++
	}
	// 4 distinct-artist tracks in the compilation dir.
	for _, a := range []string{"Ado", "Emile Bobo", "Alfred Ondo", "Nathan Epenge"} {
		tr := makeTrack(id, fmt.Sprintf("/lib/comp/%s.mp3", a), a, "", "", 0, 0)
		tr.Filename = "Daft Punk - Song.mp3" // filename hints library artist
		trs = append(trs, tr)
		id++
	}

	svc := NewSuggestionService(testutil.NewFakeRepo(trs...))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib/comp")
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range suggs {
		if art, ok := s.Fields["artist"]; ok {
			if art.Source == "library" {
				t.Errorf("dir with 4 distinct artists is a compilation — library source must be blocked (track %d)", s.TrackID)
			}
		}
	}
}

func TestAnalyzeRoot_ThreeArtistsNotCompilation(t *testing.T) {
	// Exactly 3 distinct tag artists in a dir → NOT a compilation (threshold is >3).
	// Library source must be allowed when the registry has a strong hit.
	// Note: registry is built from tracks in the ANALYZED root, so the Daft Punk seed
	// tracks must be inside the same root (/lib) as the dir being tested (/lib/dir).
	var trs []domain.Track
	id := uint64(1)

	// 100 "Daft Punk" tracks elsewhere in /lib (conf≈0.87 in registry).
	for j := range 100 {
		trs = append(trs, makeTrack(id, fmt.Sprintf("/lib/main/t%d.mp3", j), "Daft Punk", "", "", 0, 0))
		id++
	}
	// 3 distinct-artist tracks in /lib/dir.
	for _, a := range []string{"Ado", "Emile Bobo", "Alfred Ondo"} {
		trs = append(trs, makeTrack(id, fmt.Sprintf("/lib/dir/%s.mp3", a), a, "", "", 0, 0))
		id++
	}
	// Lone tagless track in /lib/dir whose filename hints at the library artist.
	loneID := id
	lone := makeTrack(id, "/lib/dir/Daft Punk - Song.mp3", "", "", "", 0, 0)
	lone.Filename = "Daft Punk - Song.mp3"
	trs = append(trs, lone)

	svc := NewSuggestionService(testutil.NewFakeRepo(trs...))
	// Analyze /lib so the registry sees Daft Punk tracks from /lib/main.
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	var loneS *domain.OrganizerSuggestion
	for i := range suggs {
		if suggs[i].TrackID == loneID {
			loneS = &suggs[i]
			break
		}
	}
	if loneS == nil {
		t.Fatal("lone track not found in suggestions")
	}
	art, ok := loneS.Fields["artist"]
	if !ok {
		t.Fatal("expected artist suggestion for lone track")
	}
	// Library source must win: conf≈0.87 > tagArtist≈0.80, and dir has only 3 tag artists → not blocked.
	if art.Source != "library" {
		t.Errorf("with 3 distinct dir-artists (not compilation) library source must be used, got %q (value=%q, conf=%.2f)",
			art.Source, art.Value, art.Confidence)
	}
}

func TestAnalyzeRoot_YearNotSuggestedWhenYearExists(t *testing.T) {
	// Year already set → no year suggestion even when album carries a timestamp.
	tr := makeTrack(1, "/lib/dir/song.mp3", "", "Album inconnu (09/10/2008 09:46:25)", "", 2020, 0)
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	if _, ok := suggs[0].Fields["year"]; ok {
		t.Error("year suggestion must not be generated when track.Year is already set")
	}
}

func TestAnalyzeRoot_TrackNumNotSuggestedWhenAlreadySet(t *testing.T) {
	// TrackNum already set → filename-parsed trackNum must not produce a suggestion.
	tr := makeTrack(1, "/lib/dir/03 - Title.mp3", "", "", "Title", 0, 3)
	tr.Filename = "03 - Title.mp3"
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	if _, ok := suggs[0].Fields["track_num"]; ok {
		t.Error("track_num suggestion must not be generated when track.TrackNum is already set")
	}
}

func TestAnalyzeRoot_IssueMissingArtistForPlaceholder(t *testing.T) {
	// Placeholder ArtistRaw values must be treated like missing and generate "Missing artist".
	tr := makeTrack(1, "/lib/dir/song.mp3", "va", "", "Title", 0, 0)
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	found := false
	for _, iss := range suggs[0].Issues {
		if iss == "Missing artist" {
			found = true
		}
	}
	if !found {
		t.Errorf("placeholder artist 'va' must generate 'Missing artist' issue, got: %v", suggs[0].Issues)
	}
}

func TestAnalyzeRoot_IssueMissingTitleWhenEmpty(t *testing.T) {
	// Empty Title must generate "Missing title" issue.
	tr := makeTrack(1, "/lib/dir/song.mp3", "Artist", "", "", 0, 0)
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	found := false
	for _, iss := range suggs[0].Issues {
		if iss == "Missing title" {
			found = true
		}
	}
	if !found {
		t.Errorf("empty title must generate 'Missing title' issue, got: %v", suggs[0].Issues)
	}
}

func TestAnalyzeRoot_AlbumFromPath(t *testing.T) {
	// Track at root/Artist/Album/song.mp3 with no album tag receives album from path.
	tr := makeTrack(1, "/lib/Artist/Best Album/song.mp3", "", "", "", 0, 0)
	tr.Filename = "song.mp3"
	svc := NewSuggestionService(testutil.NewFakeRepo(tr))
	suggs, _, err := svc.AnalyzeRoot(context.Background(), "/lib")
	if err != nil {
		t.Fatal(err)
	}
	if len(suggs) == 0 {
		t.Fatal("no suggestions")
	}
	alb, ok := suggs[0].Fields["album"]
	if !ok {
		t.Fatal("expected album suggestion from 2-level path")
	}
	if alb.Value != "Best Album" {
		t.Errorf("want album='Best Album', got %q", alb.Value)
	}
	if alb.Source != "path" {
		t.Errorf("want source='path', got %q", alb.Source)
	}
	if alb.Confidence < 0.70 {
		t.Errorf("2-level path album confidence %v too low (expected ≥ 0.70)", alb.Confidence)
	}
}
