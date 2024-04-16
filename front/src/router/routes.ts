import {
  RouteRecordRaw,
  RouteLocationNormalized,
  RouteParams,
} from 'vue-router';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Constructor<T = any> = new (...args: any[]) => T;
type TransformMap<T> = {
  [K in keyof T]: T[K] extends Constructor<infer S> ? S : never;
};

const transform =
  <T extends Record<string, Constructor>>(fields: T) =>
  (to: RouteLocationNormalized) =>
    Object.entries(to.params)
      .map(([key, value]) => ({
        [key]: (fields[key]
          ? new fields[key](value)
          : value) as TransformMap<T>[typeof key],
      }))
      .reduce((acc, cur) => ({ ...acc, ...cur }), {}) as TransformMap<T> &
      RouteParams;

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('layouts/MainLayout.vue'),
    children: [
      {
        path: '',
        name: 'login',
        component: () => import('pages/LoginView.vue'),
      },
      {
        path: '/index',
        component: () => import('pages/IndexView.vue'),
      },
      // 添加新的路由
      {
        path: '/bot-detail/:selfId',
        name: 'BotDetail',
        component: () => import('pages/BotDetailView.vue'),
        props: true, // 启用 props 将路由参数传递到组件
      },
    ],
  },

  // Always leave this as last one,
  // but you can also remove it
  {
    path: '/:catchAll(.*)*',
    component: () => import('src/pages/NotFoundView.vue'),
  },
];

export default routes;
