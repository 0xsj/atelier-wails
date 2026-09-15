import { createRawSnippet, mount, tick, unmount } from 'svelte';
import { describe, expect, it, vi } from 'vitest';
import { Button } from './primitives';
import { ActivityRail, WorkbenchShell } from '../workbench/ui/shell';
import { WorkspacePicker } from '../features/workspace';
import { createPreviewWorkspaceTransport } from '../platform/preview/workspace';
import { createWorkspaceOpenService } from '../services/workspace';

function raw(html: string) {
	return createRawSnippet(() => ({ render: () => html }));
}

function target(): HTMLDivElement {
	const element = document.createElement('div');
	document.body.append(element);
	return element;
}

describe('rendered component contracts', () => {
	it('renders a button contract and handles a click', () => {
		const host = target();
		const onClick = vi.fn();
		const instance = mount(Button, {
			target: host,
			props: { variant: 'primary', size: 'lg', children: raw('<span>Save</span>'), onclick: onClick }
		});

		const button = host.querySelector('button');
		expect(button).toBeInstanceOf(HTMLButtonElement);
		expect(button?.textContent?.trim()).toBe('Save');
		expect(button?.dataset.variant).toBe('primary');
		expect(button?.dataset.size).toBe('lg');

		button?.click();
		expect(onClick).toHaveBeenCalledOnce();

		unmount(instance);
		host.remove();
	});

	it('renders the shell regions and accessibility landmarks', () => {
		const host = target();
		const instance = mount(WorkbenchShell, {
			target: host,
			props: {
				title: 'Atelier',
				windowActive: false,
				activity: raw('<span>Activity</span>'),
				sidebar: raw('<span>Sidebar</span>'),
				main: raw('<span data-testid="editor">Editor</span>'),
				status: raw('<span>Status</span>')
			}
		});

		const shell = host.querySelector('.atelier-workbench-shell');
		expect(shell?.getAttribute('data-sidebar')).toBe('visible');
		expect(shell?.getAttribute('data-window-active')).toBe('false');
		expect(host.querySelector('[aria-label="Activity rail"]')).not.toBeNull();
		expect(host.querySelector('[aria-label="Sidebar"]')).not.toBeNull();
		expect(host.querySelector('[data-testid="editor"]')?.textContent).toBe('Editor');
		expect(host.querySelector('.atelier-workbench__status')?.textContent).toContain('Status');

		unmount(instance);
		host.remove();
	});

	it('updates the active activity item from click and keyboard input', async () => {
		const host = target();
		const onSelect = vi.fn();
		const items = [
			{ id: 'files', label: 'Files' },
			{ id: 'search', label: 'Search' }
		] as const;
		const instance = mount(ActivityRail, {
			target: host,
			props: { items, activeId: 'files', icon: raw('<span>•</span>'), onSelect }
		});

		const tabs = host.querySelectorAll<HTMLButtonElement>('[role="tab"]');
		expect(tabs).toHaveLength(2);
		expect(tabs[0]?.getAttribute('aria-selected')).toBe('true');

		tabs[1]?.click();
		await tick();
		expect(onSelect).toHaveBeenCalledWith(items[1]);
		expect(tabs[1]?.getAttribute('aria-selected')).toBe('true');

		tabs[1]?.focus();
		tabs[1]?.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowUp', bubbles: true }));
		await tick();
		expect(document.activeElement).toBe(tabs[0]);
		expect(tabs[0]?.getAttribute('aria-selected')).toBe('true');

		unmount(instance);
		host.remove();
	});

	it('renders the workspace picker and opens a selected registry record', async () => {
		const host = target();
		const onOpen = vi.fn();
		const transport = createPreviewWorkspaceTransport();
		const instance = mount(WorkspacePicker, {
			target: host,
			props: { transport, openPort: createWorkspaceOpenService(transport), storeLabel: 'Preview registry', onOpen }
		});

		await expect.poll(() => host.querySelector('[aria-label="Registered workspaces"]')).not.toBeNull();
		const filter = host.querySelector<HTMLSelectElement>('[aria-label="Workspace filter"]');
		expect(filter?.value).toBe('active');
		if (filter) {
			filter.value = 'all';
			filter.dispatchEvent(new Event('change', { bubbles: true }));
		}
		await expect.poll(() => host.querySelectorAll('[role="option"]')).toHaveLength(2);
		expect(host.textContent).toContain('Field Notes');
		const first = host.querySelector<HTMLButtonElement>('[role="option"]');
		expect(first?.textContent).toContain('Atelier Core');
		first?.click();
		await expect.poll(() => [...host.querySelectorAll('button')].find((button) => button.textContent?.includes('Open workspace'))).toBeDefined();
		const openButton = [...host.querySelectorAll('button')].find((button) => button.textContent?.includes('Open workspace'));
		openButton?.click();
		expect(onOpen).toHaveBeenCalledWith(expect.objectContaining({ scope: 'workspace:01900000-0000-7000-8000-000000000101', workspace: expect.objectContaining({ name: 'Atelier Core' }) }));

		unmount(instance);
		host.remove();
	});
});
