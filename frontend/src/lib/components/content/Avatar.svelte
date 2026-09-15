<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';
	import './content.css';

	type AvatarSize = 'sm' | 'md' | 'lg';
	type AvatarStatus = 'online' | 'away' | 'busy' | 'offline';

	interface Props extends Omit<HTMLAttributes<HTMLSpanElement>, 'class'> {
		name: string;
		src?: string;
		size?: AvatarSize;
		status?: AvatarStatus;
		class?: string;
	}

	let { name, src, size = 'md', status, class: className, ...rest }: Props = $props();
	const initials = $derived(getInitials(name));
	const accessibleLabel = $derived(status ? `${name}, ${status}` : name);

	function getInitials(value: string): string {
		return value.split(/\s+/).filter(Boolean).slice(0, 2).map((part) => part[0]).join('').toUpperCase() || '?';
	}
</script>

<span {...rest} class="atelier-avatar {className ?? ''}" data-size={size} data-status={status} role="img" aria-label={accessibleLabel}>
	{#if src}<img src={src} alt="" />{:else}<span aria-hidden="true">{initials}</span>{/if}
	{#if status}<span class="atelier-avatar__status" aria-hidden="true"></span>{/if}
</span>
