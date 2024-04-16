<template>
  <q-page padding>
    <div class="text-h6 q-mb-md">机器人管理</div>
    <q-btn icon="refresh" color="primary" @click="fetchRobots" class="q-mb-md"
      >刷新机器人状态</q-btn
    >
    <div v-if="loading">加载中...</div>
    <div v-else-if="robots.length === 0">暂无机器人接入</div>
    <div v-else>
      <q-list bordered separator>
        <q-item
          v-for="robot in robots"
          :key="robot.self_id"
          clickable
          @click="navigateToRobot(robot.self_id)"
        >
          <q-item-section avatar>
            <q-icon
              :name="robot.isOnline ? 'check_circle' : 'cancel'"
              :color="robot.isOnline ? 'green' : 'red'"
            />
          </q-item-section>
          <q-item-section avatar>
            <q-avatar>
              <img :src="'data:image/png;base64,' + robot.imgHead" />
            </q-avatar>
          </q-item-section>
          <q-item-section class="col">
            <div class="row">
              <div class="col-6">
                <div class="text-h6">{{ robot.nickname }}</div>
                <div>账号: {{ robot.self_id }}</div>
                <div>收信息数: {{ robot.message_received }}</div>
                <div>发信息数: {{ robot.message_sent }}</div>
              </div>
              <div class="col-6">
                <div>上次发信息: {{ formatDate(robot.last_message_time) }}</div>
                <div>收到邀请: {{ robot.invites_received }}</div>
                <div>被踢次数: {{ robot.kicks_received }}</div>
                <div>日活DAU: {{ robot.daily_dau }}</div>
              </div>
            </div>
          </q-item-section>
        </q-item>
      </q-list>
    </div>
  </q-page>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import {
  QPage,
  QList,
  QItem,
  QItemSection,
  QAvatar,
  QIcon,
  QBtn,
} from 'quasar';

const robots = ref([]);
const loading = ref(false);
const router = useRouter();

onMounted(() => {
  fetchRobots();
});

function fetchRobots() {
  loading.value = true;
  fetch('/webui/api/online-robots')
    .then((response) => response.json())
    .then((data) => {
      robots.value = data;
      loading.value = false;
    })
    .catch((error) => {
      console.error('Error fetching robots:', error);
      loading.value = false;
    });
}

function navigateToRobot(selfId) {
  console.log('Attempting to navigate to BotDetail with selfId:', selfId);
  router.push({ name: 'BotDetail', params: { selfId } });
}

function formatDate(timestamp) {
  return new Date(timestamp * 1000).toLocaleString();
}
</script>

<style scoped>
q-avatar img {
  border-radius: 4px;
}
</style>
