import { HomeLayout } from 'fumadocs-ui/layouts/home';
import { baseOptions, linkItems } from '@/components/layouts/shared';

export default function Layout({ children }: LayoutProps<'/[lang]'>) {
  return (
    <HomeLayout
      {...baseOptions()}
      links={linkItems}
      // 랜딩은 디자인 소스(MZ8Ua)에 맞춰 항상 다크.
      className="dark bg-[#0a0a0a] [--color-fd-background:#0a0a0a] [--color-fd-primary:var(--color-brand)]"
    >
      {children}
    </HomeLayout>
  );
}
