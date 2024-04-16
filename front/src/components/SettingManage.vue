<template>
  <q-page padding>
    <div class="text-h6 q-mb-md">配置修改</div>
    <q-btn color="primary" @click="saveConfig" class="q-mb-md">保存配置</q-btn>

    <q-form @submit.prevent="saveConfig" class="q-gutter-md">
      <q-input
        filled
        v-model="config.account"
        label="登录用户名"
        hint="输入登录用户名"
      />
      <q-input
        filled
        v-model="config.password"
        type="password"
        label="登录密码"
        hint="输入登录密码"
      />
      <q-input
        filled
        v-model="config.title"
        label="自定义标题"
        hint="输入自定义标题"
      />
      <q-input
        filled
        v-model="config.wsPath"
        label="WebSocket路径"
        hint="输入WebSocket的监听路径"
      />
      <q-input
        filled
        v-model="config.port"
        label="WebUI端口"
        hint="输入WebUI监听的端口号"
      />
      <q-toggle filled v-model="config.useHttps" label="使用 HTTPS" />
      <q-input
        filled
        v-model="config.cert"
        label="证书路径"
        hint="输入HTTPS证书路径"
      />
      <q-input
        filled
        v-model="config.key"
        label="密钥路径"
        hint="输入HTTPS密钥路径"
      />
      <q-toggle
        filled
        v-model="config.enableWsServer"
        label="启用正向WS服务器"
      />
      <q-input
        filled
        v-model="config.wsServerToken"
        label="正向WS的Token"
        hint="输入正向WebSocket服务器的Token"
      />

      <div>
        <div class="text-h6 q-mt-md">API监测信息</div>
        <div v-for="(api, index) in config.apis" :key="index">
          <q-input
            filled
            v-model="api.apiPaths"
            label="API地址"
            hint="输入API的检测存活地址"
          />
          <q-input
            filled
            v-model="api.apiNames"
            label="API名称"
            hint="输入API的名称"
          />
          <q-btn
            flat
            icon="delete"
            color="negative"
            @click="deleteApiInfo(index)"
            label="删除"
          />
        </div>
        <q-btn
          flat
          icon="add"
          color="positive"
          @click="addApiInfo"
          label="添加API信息"
        />
      </div>

      <div>
        <div class="text-h6 q-mt-md">机器人信息</div>
        <div v-for="(bot, index) in config.botInfos" :key="bot.botId">
          <q-input
            filled
            v-model="bot.botId"
            label="机器人ID"
            hint="输入机器人的唯一标识"
          />
          <q-input
            filled
            v-model="bot.botNickname"
            label="机器人昵称"
            hint="输入机器人的昵称"
          />
          <q-input
            filled
            v-model="bot.botHead"
            label="机器人头像路径"
            hint="输入机器人头像的本地路径"
          />
          <q-btn
            flat
            icon="delete"
            color="negative"
            @click="deleteBotInfo(index)"
            label="删除"
          />
        </div>

        <q-btn
          flat
          icon="add"
          color="positive"
          @click="addBotInfo"
          label="添加机器人"
        />
      </div>
    </q-form>
  </q-page>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { QPage, QBtn, QForm, QInput, QToggle } from 'quasar';

const config = ref({
  account: '',
  password: '',
  title: '',
  wsPath: '',
  port: '',
  useHttps: false,
  cert: '',
  key: '',
  enableWsServer: false,
  wsServerToken: '',
  botInfos: [],
  apis: [],
});

onMounted(() => {
  fetchConfig();
});

const fetchConfig = async () => {
  try {
    const response = await fetch('/webui/api/getjson');
    if (response.ok) {
      config.value = await response.json();
    } else {
      throw new Error('Failed to fetch config');
    }
  } catch (error) {
    console.error('Error:', error);
  }
};

const saveConfig = async () => {
  try {
    const response = await fetch('/webui/api/savejson', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(config.value),
    });
    if (response.ok) {
      this.$q.notify({
        color: 'green',
        position: 'top',
        message: '配置修改成功',
        timeout: 3000,
      });
    } else {
      throw new Error('Failed to save config');
    }
  } catch (error) {
    this.$q.notify({
      color: 'red',
      position: 'top',
      message: '配置保存失败',
      timeout: 3000,
    });
    console.error('Error:', error);
  }
};

const addBotInfo = () => {
  config.value.botInfos.push({ botId: '', botNickname: '', botHead: '' });
};

const deleteBotInfo = (index) => {
  config.value.botInfos.splice(index, 1);
};

const addApiInfo = () => {
  config.value.apis.push({ apiPaths: '', apiNames: '' });
};

const deleteApiInfo = (index) => {
  config.value.apis.splice(index, 1);
};
</script>
