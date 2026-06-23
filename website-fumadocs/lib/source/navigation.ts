export function getSection(path: string | undefined) {
  if (!path) return 'overview';

  const [dir] = path.split('/', 1);
  if (!dir) return 'overview';

  return (
    {
      'getting-started': 'getting-started',
      guide: 'guide',
      reference: 'reference',
    }[dir] ?? 'overview'
  );
}
