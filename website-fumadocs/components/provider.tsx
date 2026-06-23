'use client';
import dynamic from 'next/dynamic';
import { useEffect, type ReactNode } from 'react';
import { RootProvider } from 'fumadocs-ui/provider/base';
import { useSearchContext } from 'fumadocs-ui/contexts/search';

type I18nProp = React.ComponentProps<typeof RootProvider>['i18n'];

const SearchDialog = dynamic(() => import('@/components/search'), {
  ssr: false,
});

function SearchHotKey() {
  const { setOpenSearch, enabled } = useSearchContext();

  useEffect(() => {
    if (!enabled) return;

    const onKeyDown = (event: KeyboardEvent) => {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault();
        setOpenSearch(true);
      }
    };

    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [enabled, setOpenSearch]);

  return null;
}

export function Provider({ children, i18n }: { children: ReactNode; i18n?: I18nProp }) {
  return (
    <RootProvider search={{ SearchDialog }} i18n={i18n}>
      <SearchHotKey />
      {children}
    </RootProvider>
  );
}
