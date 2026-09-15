<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Tabs } from '../../../components/navigation';
	import './view-host.css';

	export type WorkbenchView = {
		id: string;
		label: string;
		disabled?: boolean;
	};

	interface Props {
		views: readonly WorkbenchView[];
		activeView?: string;
		children?: Snippet;
		class?: string;
	}

	let { views, activeView = $bindable(''), children, class: className }: Props = $props();
	const tabItems = $derived(views.map((view) => ({ value: view.id, label: view.label, disabled: view.disabled })));
</script>

<Tabs items={tabItems} bind:value={activeView} class="atelier-view-host {className ?? ''}">
	{@render children?.()}
</Tabs>
