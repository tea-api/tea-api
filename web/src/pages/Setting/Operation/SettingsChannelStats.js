import React, { useEffect, useState } from 'react';
import { Table, Typography, Spin, Card, Progress } from '@douyinfe/semi-ui';
import { API, showError } from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function SettingsChannelStats() {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState([]);

  const fetchStats = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/channel/stats');
      if (res.data.success) {
        // 确保data是数组
        const data = Array.isArray(res.data.data) ? res.data.data : [];
        setStats(data);
      } else {
        showError(t(res.data.message));
      }
    } catch (error) {
      showError(t('获取线路监控数据失败'));
      console.error('Error fetching channel stats:', error);
      // 发生错误时设置为空数组
      setStats([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStats();
    // 每60秒自动刷新一次数据
    const interval = setInterval(fetchStats, 60000);
    return () => clearInterval(interval);
  }, []);

  const columns = [
    {
      title: t('ID'),
      dataIndex: 'id',
      width: 80,
    },
    {
      title: t('渠道ID'),
      dataIndex: 'channel_id',
      width: 100,
    },
    {
      title: t('渠道名称'),
      dataIndex: 'name',
      width: 200,
    },
    {
      title: t('总请求数'),
      dataIndex: 'total',
      width: 120,
    },
    {
      title: t('成功请求数'),
      dataIndex: 'success',
      width: 120,
    },
    {
      title: t('成功率'),
      dataIndex: 'success_rate',
      width: 200,
      render: (_, record) => {
        const successRate = record.total > 0 ? (record.success / record.total) * 100 : 0;
        const formattedRate = successRate.toFixed(2);
        let color = 'green';
        if (successRate < 90) color = 'yellow';
        if (successRate < 70) color = 'red';
        
        return (
          <div>
            <Progress percent={successRate} stroke={color} showInfo={false} />
            <Typography.Text>{formattedRate}%</Typography.Text>
          </div>
        );
      },
    },
    {
      title: t('最后更新时间'),
      dataIndex: 'updated_at',
      width: 180,
      render: (text) => {
        const date = new Date(text);
        return date.toLocaleString();
      },
    },
  ];

  return (
    <Card
      title={t('线路监控')}
      bordered={false}
      extra={
        <Typography.Text link onClick={fetchStats}>
          {t('刷新')}
        </Typography.Text>
      }
    >
      <Spin spinning={loading}>
        <Table
          columns={columns}
          dataSource={stats || []}
          pagination={false}
          rowKey="id"
          empty={t('暂无数据')}
        />
      </Spin>
    </Card>
  );
} 