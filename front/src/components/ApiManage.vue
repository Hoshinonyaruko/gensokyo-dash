<template>
  <q-page class="q-pa-md">
    <div class="row justify-between q-mb-md">
      <q-btn icon="refresh" label="刷新" @click="fetchData" />
      <q-select
        v-model="selectedDays"
        :options="dayOptions"
        label="选择时间范围"
        style="width: auto; min-width: 250px"
        @update:model-value="fetchData"
      />
    </div>

    <q-table
      :rows="apiStatuses"
      :columns="columns"
      row-key="date"
      binary-state-sort
      flat
      bordered
      :rows-per-page-options="[7, 15, 30]"
    >
      <template v-slot:body-cell-online="{ props }">
        <q-td :props="props">
          <q-icon
            :name="props.row.online ? 'check_circle' : 'cancel'"
            :color="props.row.online ? 'green' : 'red'"
          />
        </q-td>
      </template>
    </q-table>
  </q-page>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { QPage, QBtn, QSelect, QTable, QTd, QIcon } from 'quasar';

const selectedDays = ref(7);
const dayOptions = ref([
  { label: '7天', value: 7 },
  { label: '15天', value: 15 },
  { label: '30天', value: 30 },
]);

const apiStatuses = ref([]);

const columns = ref([
  { name: 'apiPaths', label: 'API 路径', field: 'apiPaths', sortable: true },
  { name: 'apiNames', label: 'API 名称', field: 'apiNames', sortable: true },
  {
    name: 'online',
    label: '在线状态',
    field: 'online',
    sortable: true,
    align: 'center',
  },
  {
    name: 'responseTime',
    label: '响应成功次数',
    field: 'responseTime',
    sortable: true,
  },
  {
    name: 'checksPerformed',
    label: '检查次数',
    field: 'checksPerformed',
    sortable: true,
  },
  {
    name: 'checksFailed',
    label: '检查失败次数',
    field: 'checksFailed',
    sortable: true,
  },
  {
    name: 'successRate',
    label: '成功率(%)',
    field: 'successRate',
    sortable: true,
    format: (val) => val.toFixed(2),
  },
  { name: 'date', label: '日期', field: 'date', sortable: true },
]);

function fetchData() {
  // 假设后端接口已经根据天数进行过滤
  fetch(`/webui/api/api-info?days=${selectedDays.value}`)
    .then((response) => response.json())
    .then((data) => {
      apiStatuses.value = data;
    })
    .catch((error) => {
      console.error('Error fetching API statuses:', error);
    });
}

onMounted(fetchData);
</script>
