import { onDestroy, onMount } from "svelte";
import { t } from "../stores/i18n";

export type TranslateFn = (key: any) => string;

/** Reactive translation fn for Svelte 5 runes components. */
export function subscribeI18n(setTr: (fn: TranslateFn) => void): void {
  onMount(() => {
    const unsub = t.subscribe((fn) => {
      setTr(fn as TranslateFn);
    });
    onDestroy(unsub);
  });
}
