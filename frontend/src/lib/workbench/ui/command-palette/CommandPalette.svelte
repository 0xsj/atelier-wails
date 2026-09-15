<script lang="ts">
	import { QuickPick, type QuickPickItem } from '../quick-pick';

	export type CommandItem = {
		id: string;
		label: string;
		description?: string;
		shortcut?: string;
		disabled?: boolean;
	};

	interface Props {
		commands: readonly CommandItem[];
		open?: boolean;
		onExecute?: (command: CommandItem) => void;
		class?: string;
	}

	let { commands, open = $bindable(false), onExecute, class: className }: Props = $props();
	const items = $derived<readonly QuickPickItem[]>(commands.map((command) => ({ ...command })));

	function execute(item: QuickPickItem): void {
		const command = commands.find((entry) => entry.id === item.id);
		if (command) onExecute?.(command);
	}
</script>

<QuickPick bind:open {items} title="Command palette" placeholder="Search commands" emptyText="No commands match" onSelect={execute} class={className} />
