<template>
  <q-page>
    <div class="q-pa-md">
      <div class="row items-center justify-between">
        <div class="flex items-center q-gutter-sm">
          <!-- 返回按钮 -->
          <q-btn
            flat
            icon="arrow_back"
            @click="navigateToIndexWithBotStatus"
            aria-label="返回"
          />

          <!-- 刷新按钮 -->
          <q-btn icon="refresh" label="刷新" @click="fetchData" />
        </div>

        <!-- 时间选择器，放在右侧 -->
        <q-select
          v-model="selectedDays"
          :options="dayOptions"
          label="选择时间范围"
          style="width: 200px"
          @update:model-value="fetchData"
          class="q-ml-md"
        />
      </div>

      <div class="row q-my-md">
        <div
          v-for="status in robotStatuses"
          :key="status.date"
          class="q-pa-xs"
          :style="{ backgroundColor: status.online ? 'lightgreen' : 'grey' }"
        >
          {{ formatDate(status.date) }}
        </div>
      </div>
      <!-- 添加表格组件 -->
      <q-table
        :rows="robotStatuses"
        :columns="columns"
        row-key="date"
        binary-state-sort
        flat
        bordered
        :rows-per-page-options="[7, 15, 30]"
      >
        <template v-slot:body-cell-online="props">
          <q-td :props="props">
            <q-icon
              :name="props.row.online ? 'check_circle' : 'cancel'"
              :color="props.row.online ? 'green' : 'red'"
            />
          </q-td>
        </template>
      </q-table>
      <!-- 大文本块展示 -->
      <div class="q-mt-md">
        <q-card>
          <q-card-section>
            被邀请总数: {{ totalInvitesReceived }}
          </q-card-section>
        </q-card>
        <q-card>
          <q-card-section> 被踢总数: {{ totalKicksReceived }} </q-card-section>
        </q-card>
      </div>

      <div>
        <!-- 输入控件 -->
        <q-select
          filled
          v-model="rank"
          :options="rankOptions"
          option-value="value"
          option-label="label"
          label="选择排名"
          class="q-ma-md"
        />

        <q-input
          filled
          v-model="date"
          label="选择日期 (YYYY-MM-DD)"
          class="q-ma-md"
          readonly
          @click="showDatePicker = true"
        />
        <q-dialog v-model="showDatePicker">
          <q-date
            v-model="date"
            mask="YYYY-MM-DD"
            @change="showDatePicker = false"
          />
        </q-dialog>

        <!-- 加载更多指标的按钮 -->
        <q-btn
          color="blue"
          @click="loadMoreMetrics"
          label="查看排名"
          class="q-mt-md"
        />

        <!-- 表格显示总体和每日指令调用统计 -->
        <q-table
          v-if="commandAll.length > 0"
          :rows="commandAll"
          :columns="columnsCommand"
          row-key="selfId"
          class="q-mt-md"
          title="总指令调用统计"
        />
        <q-table
          v-if="commandDaily.length > 0"
          :rows="commandDaily"
          :columns="columnsCommand"
          row-key="selfId"
          class="q-mt-md"
          title="每日指令调用统计"
        />

        <q-table
          v-if="groupAll.length > 0"
          :rows="groupAll"
          :columns="columnsGroupAll"
          row-key="group_id"
          class="q-mt-md"
          title="总群组指令调用统计"
        />
        <q-table
          v-if="groupDaily.length > 0"
          :rows="groupDaily"
          :columns="columnsGroupDaily"
          row-key="group_id"
          class="q-mt-md"
          title="每日群组指令调用统计"
        />

        <q-table
          v-if="userAll.length > 0"
          :rows="userAll"
          :columns="columnsUserAll"
          row-key="user_id"
          class="q-mt-md"
          title="用户总统计"
        />
        <q-table
          v-if="userDaily.length > 0"
          :rows="userDaily"
          :columns="columnsUserDaily"
          row-key="user_id"
          class="q-mt-md"
          title="用户每日统计"
        />
      </div>
    </div>
  </q-page>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { QPage, QBtn, QSelect, QTable, QTd, QIcon } from 'quasar';
import { useRoute, useRouter } from 'vue-router';

const route = useRoute();
const router = useRouter();
const selfId = ref(route.params.selfId); // 从路由参数获取 selfId

const selectedDays = ref(7);
const dayOptions = ref([
  { label: '7天', value: 7 },
  { label: '15天', value: 15 },
  { label: '30天', value: 30 },
]);

const rankOptions = [
  { label: '10', value: 10 },
  { label: '20', value: 20 },
  { label: '50', value: 50 },
];

const robotStatuses = ref([]);
const onlineData = ref([]);
const messageReceivedData = ref([]);
const messageSentData = ref([]);
const dailyDAUData = ref([]);
const totalInvitesReceived = ref(0);
const totalKicksReceived = ref(0);

const columns = ref([
  {
    name: 'date',
    required: true,
    label: '日期',
    align: 'left',
    field: (row) => row.date,
    sortable: true,
  },
  { name: 'online', label: '在线状态', align: 'center', sortable: true },
  {
    name: 'message_received',
    label: '收到的消息数',
    field: 'message_received',
    sortable: true,
  },
  {
    name: 'message_sent',
    label: '发送的消息数',
    field: 'message_sent',
    sortable: true,
  },
  {
    name: 'last_message_time',
    label: '最后消息时间',
    field: 'last_message_time',
    sortable: true,
  },
  {
    name: 'invites_received',
    label: '收到的邀请数',
    field: 'invites_received',
    sortable: true,
  },
  {
    name: 'kicks_received',
    label: '被踢次数',
    field: 'kicks_received',
    sortable: true,
  },
  {
    name: 'daily_dau',
    label: '日活跃用户数',
    field: 'daily_dau',
    sortable: true,
  },
]);

const rank = ref(10);
const date = ref('');

const commandAll = ref([]);
const commandDaily = ref([]);

const groupAll = ref([]);
const groupDaily = ref([]);

const userAll = ref([]);
const userDaily = ref([]);

const showDatePicker = ref(false);

const columnsCommand = [
  {
    name: 'command_name',
    label: '指令名称',
    field: 'command_name',
    sortable: true,
  },
  {
    name: 'total_calls',
    label: '总调用次数',
    field: 'total_calls',
    sortable: true,
  },
  {
    name: 'last_call_timestamp',
    label: '最后调用时间',
    field: 'last_call_timestamp',
    sortable: true,
    format: (val) => new Date(val * 1000).toLocaleString(),
  },
];

const columnsGroupAll = [
  { name: 'groupID', label: '群组ID', field: 'group_id', sortable: true },
  { name: 'selfID', label: '自身ID', field: 'self_id', sortable: true },
  {
    name: 'totalMessagesSent',
    label: '总消息发送数',
    field: 'total_messages_sent',
    sortable: true,
  },
  {
    name: 'lastMessageTimestamp',
    label: '最后消息时间',
    field: 'last_message_timestamp',
    format: (val) => new Date(val * 1000).toLocaleString(),
    sortable: true,
  },
  {
    name: 'consecutiveMessageDays',
    label: '连续消息天数',
    field: 'consecutive_message_days',
    sortable: true,
  },
];

const columnsGroupDaily = [
  { name: 'date', label: '日期', field: 'date', sortable: true },
  {
    name: 'messagesSent',
    label: '发送消息数',
    field: 'messages_sent',
    sortable: true,
  },
  {
    name: 'activeMembers',
    label: '活跃成员数',
    field: 'active_members',
    sortable: true,
  },
];

const columnsUserAll = [
  { name: 'userID', label: '用户ID', field: 'user_id', sortable: true },
  { name: 'nickname', label: '昵称', field: 'nickname', sortable: true },
  { name: 'role', label: '角色', field: 'role', sortable: true },
  {
    name: 'totalMessagesSent',
    label: '总消息数',
    field: 'total_messages_sent',
    sortable: true,
  },
  {
    name: 'lastMessageTimestamp',
    label: '最后消息时间',
    field: 'last_message_timestamp',
    format: (val) => new Date(val * 1000).toLocaleString(),
    sortable: true,
  },
  {
    name: 'consecutiveMessageDays',
    label: '连续消息天数',
    field: 'consecutive_message_days',
    sortable: true,
  },
];

const columnsUserDaily = [
  { name: 'date', label: '日期', field: 'date', sortable: true },
  {
    name: 'messagesSent',
    label: '消息数',
    field: 'messages_sent',
    sortable: true,
  },
  {
    name: 'activeMembers',
    label: '活跃成员',
    field: 'active_members',
    sortable: true,
  },
  {
    name: 'includedInGroupCount',
    label: '群组计数包含',
    field: 'included_in_group_count',
    format: (val) => (val ? '是' : '否'),
    sortable: true,
  },
];

function loadMoreMetrics() {
  if (!rank.value || !date.value) {
    console.error('Date and rank are required.');
    return;
  }

  Promise.all([
    fetch(`/webui/api/command-all?rank=${rank.value}`).then((res) =>
      res.json()
    ),
    fetch(
      `/webui/api/command-daily?date=${date.value}&rank=${rank.value}`
    ).then((res) => res.json()),
    fetch(`/webui/api/group-all?rank=${rank.value}`).then((res) => res.json()),
    fetch(`/webui/api/group-daily?date=${date.value}&rank=${rank.value}`).then(
      (res) => res.json()
    ),
    fetch(`/webui/api/user-all?rank=${rank.value}`).then((res) => res.json()),
    fetch(`/webui/api/user-daily?date=${date.value}&rank=${rank.value}`).then(
      (res) => res.json()
    ),
  ])
    .then(
      ([
        allData,
        dailyData,
        groupAllData,
        groupDailyData,
        userAllData,
        userDailyData,
      ]) => {
        commandAll.value = allData;
        commandDaily.value = dailyData;
        groupAll.value = groupAllData;
        groupDaily.value = groupDailyData;
        userAll.value = userAllData;
        userDaily.value = userDailyData;
      }
    )
    .catch((error) => {
      console.error('Failed to fetch data:', error);
    });
}

// 从BotDetailView.vue
function navigateToIndexWithBotStatus() {
  router.push({ path: '/index', query: { tab: 'botstatus' } });
}

function formatDate(dateString) {
  return new Date(dateString).toLocaleDateString(); // 格式化日期为 YYYY/MM/DD
}

function fetchData() {
  const url = `/webui/api/robot-info-all?selfID=${selfId.value}&days=${selectedDays.value}`;
  fetch(url)
    .then((response) => response.json())
    .then((data) => {
      robotStatuses.value = data;
      processRobotData(data);
    })
    .catch((error) => console.error('Error fetching robot data:', error));
}

function processRobotData(data) {
  onlineData.value = data.map((status) => ({
    date: status.date,
    color: status.online ? 'green' : 'grey',
  }));
  messageReceivedData.value = data.map((status) => ({
    date: status.date,
    value: status.message_received,
  }));
  messageSentData.value = data.map((status) => ({
    date: status.date,
    value: status.message_sent,
  }));
  dailyDAUData.value = data.map((status) => ({
    date: status.date,
    value: status.daily_dau,
  }));
  totalInvitesReceived.value = data.reduce(
    (sum, item) => sum + item.invites_received,
    0
  );
  totalKicksReceived.value = data.reduce(
    (sum, item) => sum + item.kicks_received,
    0
  );
}

onMounted(fetchData);
</script>
