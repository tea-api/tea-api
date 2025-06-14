import React, { useEffect, useState } from 'react';
import { Card, Spin, Tabs } from '@douyinfe/semi-ui';
import SettingsGeneral from '../pages/Setting/Operation/SettingsGeneral.js';
import SettingsDrawing from '../pages/Setting/Operation/SettingsDrawing.js';
import SettingsSensitiveWords from '../pages/Setting/Operation/SettingsSensitiveWords.js';
import SettingsLog from '../pages/Setting/Operation/SettingsLog.js';
import SettingsDataDashboard from '../pages/Setting/Operation/SettingsDataDashboard.js';
import SettingsMonitoring from '../pages/Setting/Operation/SettingsMonitoring.js';
import SettingsCreditLimit from '../pages/Setting/Operation/SettingsCreditLimit.js';
import ModelSettingsVisualEditor from '../pages/Setting/Operation/ModelSettingsVisualEditor.js';
import GroupRatioSettings from '../pages/Setting/Operation/GroupRatioSettings.js';
import ModelRatioSettings from '../pages/Setting/Operation/ModelRatioSettings.js';
import SettingsChannelStats from '../pages/Setting/Operation/SettingsChannelStats.js';

import { API, showError, showSuccess } from '../helpers';
import SettingsChats from '../pages/Setting/Operation/SettingsChats.js';
import { useTranslation } from 'react-i18next';
import ModelRatioNotSetEditor from '../pages/Setting/Operation/ModelRatioNotSetEditor.js';

const OperationSetting = () => {
  const { t } = useTranslation();
  let [inputs, setInputs] = useState({
    QuotaForNewUser: 0,
    QuotaForInviter: 0,
    QuotaForInvitee: 0,
    QuotaRemindThreshold: 0,
    PreConsumedQuota: 0,
    StreamCacheQueueLength: 0,
    ModelRatio: '',
    CacheRatio: '',
    CompletionRatio: '',
    ModelPrice: '',
    GroupRatio: '',
    UserUsableGroups: '',
    TopUpLink: '',
    'general_setting.docs_link': '',
    // ChatLink2: '', // 添加的新状态变量
    QuotaPerUnit: 0,
    AutomaticDisableChannelEnabled: false,
    AutomaticEnableChannelEnabled: false,
    ChannelDisableThreshold: 0,
    LogConsumeEnabled: false,
    DisplayInCurrencyEnabled: false,
    DisplayTokenStatEnabled: false,
    CheckSensitiveEnabled: false,
    CheckSensitiveOnPromptEnabled: false,
    CheckSensitiveOnCompletionEnabled: '',
    StopOnSensitiveEnabled: '',
    SensitiveWords: '',
    MjNotifyEnabled: false,
    MjAccountFilterEnabled: false,
    MjModeClearEnabled: false,
    MjForwardUrlEnabled: false,
    MjActionCheckSuccessEnabled: false,
    DrawingEnabled: false,
    DataExportEnabled: false,
    DataExportDefaultTime: 'hour',
    DataExportInterval: 5,
    DefaultCollapseSidebar: false, // 默认折叠侧边栏
    RetryTimes: 0,
    Chats: '[]',
    DemoSiteEnabled: false,
    SelfUseModeEnabled: false,
    CheckinEnabled: false, // 添加签到启用状态
    BaseCheckinReward: 10000, // 添加基础签到奖励
    ContinuousCheckinReward: 1000, // 添加连续签到奖励
    MaxContinuousRewardDays: 7, // 添加连续签到最大天数
    AutomaticDisableKeywords: '',
  });

  let [loading, setLoading] = useState(false);

  const getOptions = async () => {
    try {
      const res = await API.get('/api/option/');
      const { success, message, data } = res.data;
      if (success && data && Array.isArray(data)) {
        let newInputs = {};
        data.forEach((item) => {
          // 确保 item 存在且有 key 和 value 属性
          if (!item || typeof item.key !== 'string') {
            console.warn('Invalid option item:', item);
            return;
          }

          if (
            item.key === 'ModelRatio' ||
            item.key === 'GroupRatio' ||
            item.key === 'UserUsableGroups' ||
            item.key === 'CompletionRatio' ||
            item.key === 'ModelPrice' ||
            item.key === 'CacheRatio'
          ) {
            try {
              item.value = JSON.stringify(JSON.parse(item.value), null, 2);
            } catch (error) {
              console.error('failed to parse option', item.key, error);
              item.value = item.value || '';
            }
          }
          if (
            item.key === 'SpecialRewardDays' ||
            item.key === 'SpecialRewards'
          ) {
            try {
              item.value = JSON.parse(item.value);
              if (!Array.isArray(item.value)) {
                item.value = [];
              }
            } catch (error) {
              console.error('failed to parse option', item.key, error);
              item.value = [];
            }
          }
          if (
            item.key.endsWith('Enabled') ||
            ['DefaultCollapseSidebar'].includes(item.key)
          ) {
            newInputs[item.key] = item.value === 'true' ? true : false;
          } else {
            newInputs[item.key] = item.value;
          }
        });

        setInputs(newInputs);
      } else {
        showError(message || '获取配置失败');
        console.error('API response invalid:', { success, message, data });
      }
    } catch (error) {
      console.error('获取配置失败:', error);
      showError('获取配置失败，请刷新页面重试');
    }
  };
  async function onRefresh() {
    try {
      setLoading(true);
      await getOptions();
      // showSuccess('刷新成功');
    } catch (error) {
      showError('刷新失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    onRefresh();
  }, []);

  return (
    <>
      <Spin spinning={loading} size='large'>
        {/* 通用设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsGeneral options={inputs} refresh={onRefresh} />
        </Card>
        {/* 绘图设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsDrawing options={inputs} refresh={onRefresh} />
        </Card>
        {/* 屏蔽词过滤设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsSensitiveWords options={inputs} refresh={onRefresh} />
        </Card>
        {/* 日志设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsLog options={inputs} refresh={onRefresh} />
        </Card>
        {/* 数据看板 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsDataDashboard options={inputs} refresh={onRefresh} />
        </Card>
        {/* 监控设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsMonitoring options={inputs} refresh={onRefresh} />
        </Card>
        {/* 线路监控 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsChannelStats />
        </Card>
        {/* 额度设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsCreditLimit options={inputs} refresh={onRefresh} />
        </Card>
        {/* 聊天设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsChats options={inputs} refresh={onRefresh} />
        </Card>
        {/* 分组倍率设置 */}
        <Card style={{ marginTop: '10px' }}>
          <GroupRatioSettings options={inputs} refresh={onRefresh} />
        </Card>
        {/* 合并模型倍率设置和可视化倍率设置 */}
        <Card style={{ marginTop: '10px' }}>
          <Tabs type='line'>
            <Tabs.TabPane tab={t('模型倍率设置')} itemKey='model'>
              <ModelRatioSettings options={inputs} refresh={onRefresh} />
            </Tabs.TabPane>
            <Tabs.TabPane tab={t('可视化倍率设置')} itemKey='visual'>
              <ModelSettingsVisualEditor options={inputs} refresh={onRefresh} />
            </Tabs.TabPane>
            <Tabs.TabPane tab={t('未设置倍率模型')} itemKey='unset_models'>
              <ModelRatioNotSetEditor options={inputs} refresh={onRefresh} />
            </Tabs.TabPane>
          </Tabs>
        </Card>
      </Spin>
    </>
  );
};

export default OperationSetting;
