import { createI18nMiddleware } from 'fumadocs-core/i18n/middleware';
import { i18n } from '@/lib/i18n';

export default createI18nMiddleware(i18n);

export const config = {
  // 라우트 핸들러(api·llms·og)·정적파일·_next 는 제외하고 페이지만 로케일 처리
  matcher: ['/((?!api|og|_next|.*\\.).*)'],
};
