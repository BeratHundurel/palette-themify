export const LOCAL_ITEM_PREFIX = 'local_';

export function isLocalId(id: string): boolean {
	return id.startsWith(LOCAL_ITEM_PREFIX);
}
