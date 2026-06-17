import { writable } from "svelte/store";
import type { Track } from "../types";
import type { ArtistSuggestion } from "../types";

const duplicateGroups = writable<Track[][]>([]);
const artistSuggestions = writable<ArtistSuggestion[]>([]);
const loading = writable(false);
const error = writable<string | null>(null);
const hasRunOnce = writable(false);

/** Store for Phase 4 analysis results: duplicate groups and artist suggestions (read-only display). */
export const analysis = {
  duplicateGroups,
  artistSuggestions,
  loading,
  error,
  hasRunOnce,
  setDuplicates(groups: Track[][]) {
    duplicateGroups.set(Array.isArray(groups) ? groups : []);
  },
  setArtistSuggestions(suggestions: ArtistSuggestion[]) {
    artistSuggestions.set(Array.isArray(suggestions) ? suggestions : []);
  },
  setHasRunOnce(value: boolean) {
    hasRunOnce.set(value);
  },
  setLoading(value: boolean) {
    loading.set(value);
  },
  setError(msg: string | null) {
    error.set(msg);
  },
  removeFromGroup(trackId: number) {
    duplicateGroups.update(groups =>
      groups
        .map(g => g.filter(t => t.id !== trackId))
        .filter(g => g.length > 1)
    );
  },
  resolveGroup(firstTrackPath: string) {
    duplicateGroups.update(groups => groups.filter(g => g[0]?.path !== firstTrackPath));
  },
  reset() {
    duplicateGroups.set([]);
    artistSuggestions.set([]);
    loading.set(false);
    error.set(null);
    hasRunOnce.set(false);
  },
};
