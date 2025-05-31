import React, { useEffect, useState } from 'react';
import { Table, Tag } from '@douyinfe/semi-ui';
import { API, showError } from '../../helpers';
import { useTranslation } from 'react-i18next';

const ChannelHealth = () => {
  const { t } = useTranslation();
  const [stats, setStats] = useState([]);

  useEffect(() => {
    API.get('/api/channel/stats')
      .then((res) => {
        const { success, message, data } = res.data;
        if (success) {
          setStats(data);
        } else {
          showError(message);
        }
      })
      .catch(() => showError('Failed to load stats'));
  }, []);

  const columns = [
    { title: 'ID', dataIndex: 'channel_id' },
    { title: t('名称'), dataIndex: 'name' },
    { title: t('请求数'), dataIndex: 'total' },
    { title: t('成功数'), dataIndex: 'success' },
    {
      title: t('成功率'),
      render: (_, record) => {
        const rate = record.total
          ? ((record.success / record.total) * 100).toFixed(2) + '%'
          : '0%';
        return <Tag color='green'>{rate}</Tag>;
      },
    },
  ];

  return <Table columns={columns} dataSource={stats} rowKey='channel_id' />;
};

export default ChannelHealth;
