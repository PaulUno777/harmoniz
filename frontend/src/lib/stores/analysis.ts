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
  reset() {
    duplicateGroups.set([]);
    artistSuggestions.set([]);
    loading.set(false);
    error.set(null);
    hasRunOnce.set(false);
  },
};
