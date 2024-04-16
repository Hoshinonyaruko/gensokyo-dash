<template>
  <q-layout view="hHh lpR fFf">
    <!-- 顶部滑动栏 -->
    <q-header
      class="q-layout__section--marginal main-layout-header shadow-up-2 fixed-top"
    >
      <q-tabs v-model="tab" align="justify" scrollable>
        <q-tab name="settings" label="配置修改" />
        <q-tab name="botstatus" label="机器人监控" />
        <q-tab name="apistatus" label="API监控" />
      </q-tabs>
    </q-header>

    <!-- 主页面内容区 -->
    <q-page-container class="custom-flex-fit fit column no-wrap">
      <q-page padding v-if="tab === 'settings'">
        <!-- 配置修改页面内容 -->
        <setting-manage />
      </q-page>
      <q-page padding v-if="tab === 'botstatus'">
        <!-- 机器人监控页面内容 -->
        <bot-manage />
      </q-page>
      <q-page padding v-if="tab === 'apistatus'">
        <!-- API监控页面内容 -->
        <api-manage />
      </q-page>
    </q-page-container>
  </q-layout>
</template>

<script setup>
import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import BotManage from 'components/BotManage.vue';
import SettingManage from 'components/SettingManage.vue';
import ApiManage from 'components/ApiManage.vue';

const route = useRoute();
const tab = ref('settings'); // 默认选项卡

// 监听路由变化以更新选项卡
watch(
  () => route.query.tab,
  (newTab) => {
    if (newTab) {
      tab.value = newTab;
    }
  },
  { immediate: true }
);
</script>

<style scoped>
/* 这里可以添加一些针对不同页面的样式 */
</style>
