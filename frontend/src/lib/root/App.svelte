<script lang="ts">
  // Composition root: selects the native or preview adapters and injects
  // them. Nothing below root imports this file or the platform runtime.
  import "../styles/index.css";
  import KitchenSink from "../../dev/gallery/KitchenSink.svelte";
  import { detectHost } from "../platform/desktop/host";
  import { createDesktopPreferencesTransport } from "../platform/desktop/preferences";
  import { createDesktopWorkspaceTransport } from "../platform/desktop/workspace";
  import { createDesktopSnapshotPort } from "../platform/desktop/workbench";
  import { createPreviewPreferencesTransport } from "../platform/preview/preferences";
  import { createPreviewWorkspaceTransport } from "../platform/preview/workspace";
  import { createWorkspaceOpenService } from "../services/workspace";
  import { createMemorySnapshotPort } from "../workbench/app";

  const host = "Wails / Go";
  const dragRegion = { style: "--wails-draggable: drag" };
  const dragExclude = { style: "--wails-draggable: no-drag" };
  const hostKind = detectHost();
  const preview = hostKind === "browser" ? createPreviewPreferencesTransport() : null;
  const snapshotPort = hostKind === "browser" ? createMemorySnapshotPort() : createDesktopSnapshotPort();
  const preferences = {
    transport: preview ?? createDesktopPreferencesTransport(),
    preview,
    storeLabel: hostKind === "browser" ? "In-memory preview store; nothing persists across reloads" : "Wails bound methods over the file-backed store in the user config directory"
  };
  const workspaceTransport = hostKind === "browser" ? createPreviewWorkspaceTransport() : createDesktopWorkspaceTransport();
  const workspace = {
    transport: workspaceTransport,
    openPort: createWorkspaceOpenService(workspaceTransport),
    storeLabel: hostKind === "browser" ? "In-memory preview registry" : "Wails bound methods over persistent Workspace metadata"
  };
</script>

<KitchenSink {host} {dragRegion} {dragExclude} {preferences} workbench={{ snapshotPort }} {workspace} />
