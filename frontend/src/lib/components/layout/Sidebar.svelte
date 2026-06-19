<script lang="ts">
  import { t } from "../../stores/i18n";
  import {
    MusicNotesIcon,
    BrowserIcon,
    TrashIcon,
    GearIcon,
  } from "phosphor-svelte";
  import logoRaw from "../../../assets/images/hamonizer_logo.svg?raw";
  import { updateStore } from "../../stores/update";

  let { activeTab = $bindable("library") } = $props();

  const navItems = $derived([
    { id: "library", label: $t("library"), icon: MusicNotesIcon },
    { id: "organizer", label: $t("organizer"), icon: BrowserIcon },
    { id: "cleaner", label: $t("cleaner"), icon: TrashIcon },
  ]);
</script>

<aside class="w-52 h-full bg-surface border-r border-border flex flex-col pt-6">
  <!-- Brand -->
  <div class="px-5 mb-8 flex items-center gap-3">
    <div class="w-9 h-9 text-accent shrink-0 flex items-center justify-center">
      {@html logoRaw}
    </div>
    <div>
      <h1 class="text-base font-bold tracking-tight leading-none">Harmoniz</h1>
      <p
        class="text-[9px] text-text-muted uppercase tracking-widest font-semibold opacity-50 mt-0.5"
      >
        Bring order to the noise
      </p>
    </div>
  </div>

  <!-- Navigation -->
  <nav class="flex-1 px-3 space-y-1">
    {#each navItems as item}
      <button
        onclick={() => (activeTab = item.id)}
        class="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-150 group
               {activeTab === item.id
          ? 'bg-accent/10 text-accent'
          : 'hover:bg-white/5 text-text-secondary hover:text-text-primary'}"
      >
        <item.icon
          size={18}
          weight={activeTab === item.id ? "fill" : "regular"}
          class="shrink-0 transition-transform group-hover:scale-110"
        />
        <span class="font-medium text-sm">{item.label}</span>
      </button>
    {/each}
  </nav>

  <!-- Footer -->
  <div class="px-3 pb-4 space-y-2">
    <button
      onclick={() => (activeTab = "settings")}
      class="w-full flex items-center gap-3 px-3 py-2.5 rounded-lg text-text-secondary hover:bg-white/5 hover:text-text-primary transition-all group"
    >
      <GearIcon
        size={18}
        class="shrink-0 group-hover:rotate-45 transition-transform duration-300"
      />
      <span class="font-medium text-sm">{$t("settings")}</span>
    </button>
    <!-- version -->
    <div class="flex items-center gap-3 px-3">
      <div class="w-px h-3 bg-border shrink-0"></div>
      <div
        class="text-[9px] text-text-secondary uppercase tracking-widest font-bold opacity-60"
      >
        v{$updateStore.currentVersion}
      </div>
    </div>
  </div>
</aside>
