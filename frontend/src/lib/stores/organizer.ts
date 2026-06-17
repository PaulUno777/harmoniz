import { writable, derived } from "svelte/store";
import type { OrganizerSuggestion, FilenamePreview } from "../types";

type OrganizerStatus = "idle" | "analyzing" | "ready" | "applying";

const suggestions = writable<OrganizerSuggestion[]>([]);
const status = writable<OrganizerStatus>("idle");
const error = writable<string | null>(null);
const filenameTemplate = writable("{artist} - {title}.{ext}");
const filenamePreviews = writable<FilenamePreview[]>([]);
const lastAppliedCount = writable(0);
const selectedTrackId = writable<number | null>(null);

const withSuggestions = derived(suggestions, ($s) =>
  $s.filter((s) => Object.keys(s.fields).length > 0),
);

const selectedSuggestion = derived(
  [suggestions, selectedTrackId],
  ([$s, $id]) =>
    $id !== null ? ($s.find((s) => s.track_id === $id) ?? null) : null,
);

const highConfidenceCount = derived(
  suggestions,
  ($s) => $s.filter((s) => s.score >= 75).length,
);

const completeCount = derived(
  suggestions,
  ($s) =>
    $s.filter(
      (s) => s.issues.length === 0 && Object.keys(s.fields).length === 0,
    ).length,
);

export const organizer = {
  suggestions,
  status,
  error,
  filenameTemplate,
  filenamePreviews,
  lastAppliedCount,
  withSuggestions,
  highConfidenceCount,
  completeCount,
  selectedTrackId,
  selectedSuggestion,

  startAnalysis() {
    status.set("analyzing");
    error.set(null);
    filenamePreviews.set([]);
  },

  setResults(suggs: OrganizerSuggestion[]) {
    suggestions.set(suggs ?? []);
    status.set("ready");
  },

  setError(msg: string) {
    error.set(msg);
    status.set("idle");
  },

  startApplying() {
    status.set("applying");
  },

  setApplied(count: number) {
    lastAppliedCount.set(count);
    status.set("ready");
  },

  setPreviews(previews: FilenamePreview[]) {
    filenamePreviews.set(previews ?? []);
  },

  /** Mark an entire track as fully applied (all fields done). */
  markApplied(trackID: number) {
    suggestions.update(($s) =>
      $s.map((s) =>
        s.track_id === trackID ? { ...s, fields: {}, issues: [], score: 0 } : s,
      ),
    );
  },

  /**
   * After applying specific fields, remove them from the suggestion and update
   * the track's cached metadata so the detail panel reflects the new state.
   */
  updateTrackSuggestion(
    trackID: number,
    appliedFields: Record<string, string>,
  ) {
    suggestions.update(($s) =>
      $s.map((s) => {
        if (s.track_id !== trackID) return s;
        const fields = { ...s.fields };
        for (const key of Object.keys(appliedFields)) delete fields[key];

        // Reflect applied values on the cached track object
        const track = { ...s.track };
        if (appliedFields["artist"]) track.artist_raw = appliedFields["artist"];
        if (appliedFields["title"]) track.title = appliedFields["title"];
        if (appliedFields["album"]) track.album_raw = appliedFields["album"];
        if (appliedFields["year"])
          track.year = parseInt(appliedFields["year"]) || track.year;
        if (appliedFields["track_num"])
          track.track_num =
            parseInt(appliedFields["track_num"]) || track.track_num;

        const remaining = Object.values(fields);
        const score =
          remaining.length === 0
            ? 0
            : Math.min(
                (remaining.reduce((sum, f) => sum + f.confidence, 0) /
                  remaining.length) *
                  100,
                100,
              );

        return {
          ...s,
          fields,
          track,
          score,
          issues: remaining.length === 0 ? [] : s.issues,
        };
      }),
    );
  },

  setSelectedTrack(id: number | null) {
    selectedTrackId.set(id);
  },

  reset() {
    suggestions.set([]);
    status.set("idle");
    error.set(null);
    filenamePreviews.set([]);
    lastAppliedCount.set(0);
    selectedTrackId.set(null);
  },
};
