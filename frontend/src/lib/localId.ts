export const LOCAL_ITEM_PREFIX = 'local_';

export function isLocalId(id: string | number): boolean {
	const strId = String(id);
	return strId.startsWith(LOCAL_ITEM_PREFIX);
}
