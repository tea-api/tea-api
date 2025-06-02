import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Table, 
  Typography, 
  Space, 
  Tag, 
  Select, 
  DatePicker, 
  Button,
  Input,
  Row,
  Col
} from '@douyinfe/semi-ui';
import { IconList, IconRefresh, IconSearch } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError } from '../../helpers';

const { Title, Text } = Typography;
const { Option } = Select;

const SecurityLogs = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [logs, setLogs] = useState([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0
  });
  const [filters, setFilters] = useState({
    type: 'all',
    ip: '',
    dateRange: null
  });

  const loadLogs = async (page = 1, pageSize = 20) => {
    setLoading(true);
    try {
      const params = {
        page,
        limit: pageSize,
        type: filters.type,
        ip: filters.ip
      };

      if (filters.dateRange && filters.dateRange.length === 2) {
        params.start_date = filters.dateRange[0].toISOString();
        params.end_date = filters.dateRange[1].toISOString();
      }

      const response = await API.get('/api/security/logs', { params });
      
      if (response.data.success) {
        setLogs(response.data.data.logs || []);
        setPagination({
          current: page,
          pageSize,
          total: response.data.data.total || 0
        });
      }
    } catch (error) {
      showError(t('加载安全日志失败'));
      // 使用模拟数据
      setLogs([
        {
          id: 1,
          timestamp: '2024-01-01T12:00:00Z',
          type: 'malicious_detection',
          ip: '192.168.1.100',
          message: '检测到token浪费攻击',
          action: 'blocked',
          details: {
            prompt_length: 75000,
            random_chars: true,
            stream: true
          }
        },
        {
          id: 2,
          timestamp: '2024-01-01T11:55:00Z',
          type: 'rate_limit',
          ip: '192.168.1.101',
          message: '请求频率过高',
          action: 'rate_limited',
          details: {
            requests_per_second: 25
          }
        },
        {
          id: 3,
          timestamp: '2024-01-01T11:50:00Z',
          type: 'ip_blacklist',
          ip: '192.168.1.102',
          message: 'IP已加入黑名单',
          action: 'blacklisted',
          details: {
            reason: '恶意攻击',
            temporary: true
          }
        }
      ]);
      setPagination({
        current: 1,
        pageSize: 20,
        total: 3
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadLogs();
  }, []);

  const handleTableChange = (pagination) => {
    loadLogs(pagination.current, pagination.pageSize);
  };

  const handleSearch = () => {
    loadLogs(1, pagination.pageSize);
  };

  const handleRefresh = () => {
    loadLogs(pagination.current, pagination.pageSize);
  };

  const getTypeColor = (type) => {
    const colors = {
      malicious_detection: 'red',
      rate_limit: 'orange',
      ip_blacklist: 'purple',
      stream_protection: 'blue',
      request_limit: 'cyan',
      normal: 'green'
    };
    return colors[type] || 'grey';
  };

  const getTypeText = (type) => {
    const texts = {
      malicious_detection: t('恶意检测'),
      rate_limit: t('频率限制'),
      ip_blacklist: t('IP黑名单'),
      stream_protection: t('流保护'),
      request_limit: t('请求限制'),
      normal: t('正常')
    };
    return texts[type] || type;
  };

  const getActionColor = (action) => {
    const colors = {
      blocked: 'red',
      rate_limited: 'orange',
      blacklisted: 'purple',
      allowed: 'green',
      warning: 'yellow'
    };
    return colors[action] || 'grey';
  };

  const getActionText = (action) => {
    const texts = {
      blocked: t('已阻止'),
      rate_limited: t('已限流'),
      blacklisted: t('已封禁'),
      allowed: t('已允许'),
      warning: t('警告')
    };
    return texts[action] || action;
  };

  const columns = [
    {
      title: t('时间'),
      dataIndex: 'timestamp',
      key: 'timestamp',
      width: 180,
      render: (time) => new Date(time).toLocaleString(),
      sorter: true
    },
    {
      title: t('类型'),
      dataIndex: 'type',
      key: 'type',
      width: 120,
      render: (type) => (
        <Tag color={getTypeColor(type)}>
          {getTypeText(type)}
        </Tag>
      )
    },
    {
      title: t('IP地址'),
      dataIndex: 'ip',
      key: 'ip',
      width: 140
    },
    {
      title: t('消息'),
      dataIndex: 'message',
      key: 'message',
      ellipsis: true
    },
    {
      title: t('处理结果'),
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (action) => (
        <Tag color={getActionColor(action)}>
          {getActionText(action)}
        </Tag>
      )
    },
    {
      title: t('详情'),
      dataIndex: 'details',
      key: 'details',
      width: 200,
      render: (details) => {
        if (!details) return '-';
        return (
          <div style={{ fontSize: '12px', color: '#666' }}>
            {Object.entries(details).map(([key, value]) => (
              <div key={key}>
                {key}: {typeof value === 'boolean' ? (value ? t('是') : t('否')) : value}
              </div>
            ))}
          </div>
        );
      }
    }
  ];

  return (
    <div>
      <Card>
        <div style={{ marginBottom: '24px' }}>
          <Space>
            <IconList size="large" style={{ color: '#1890ff' }} />
            <Title heading={4} style={{ margin: 0 }}>
              {t('安全日志')}
            </Title>
          </Space>
          <Text type="secondary" style={{ display: 'block', marginTop: '8px' }}>
            {t('查看系统安全事件和处理记录')}
          </Text>
        </div>

        {/* 筛选条件 */}
        <Card style={{ marginBottom: '24px' }}>
          <Row gutter={16}>
            <Col span={6}>
              <Text strong>{t('事件类型')}</Text>
              <Select
                value={filters.type}
                onChange={(value) => setFilters({ ...filters, type: value })}
                style={{ width: '100%', marginTop: '8px' }}
              >
                <Option value="all">{t('全部')}</Option>
                <Option value="malicious_detection">{t('恶意检测')}</Option>
                <Option value="rate_limit">{t('频率限制')}</Option>
                <Option value="ip_blacklist">{t('IP黑名单')}</Option>
                <Option value="stream_protection">{t('流保护')}</Option>
                <Option value="request_limit">{t('请求限制')}</Option>
              </Select>
            </Col>
            <Col span={6}>
              <Text strong>{t('IP地址')}</Text>
              <Input
                value={filters.ip}
                onChange={(value) => setFilters({ ...filters, ip: value })}
                placeholder={t('输入IP地址')}
                style={{ marginTop: '8px' }}
              />
            </Col>
            <Col span={8}>
              <Text strong>{t('时间范围')}</Text>
              <DatePicker
                type="dateTimeRange"
                value={filters.dateRange}
                onChange={(value) => setFilters({ ...filters, dateRange: value })}
                style={{ width: '100%', marginTop: '8px' }}
              />
            </Col>
            <Col span={4}>
              <div style={{ marginTop: '32px' }}>
                <Space>
                  <Button
                    type="primary"
                    icon={<IconSearch />}
                    onClick={handleSearch}
                  >
                    {t('搜索')}
                  </Button>
                  <Button
                    icon={<IconRefresh />}
                    onClick={handleRefresh}
                  >
                    {t('刷新')}
                  </Button>
                </Space>
              </div>
            </Col>
          </Row>
        </Card>

        {/* 日志表格 */}
        <Table
          columns={columns}
          dataSource={logs}
          loading={loading}
          pagination={{
            ...pagination,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) => 
              `${t('第')} ${range[0]}-${range[1]} ${t('条，共')} ${total} ${t('条')}`
          }}
          onChange={handleTableChange}
          rowKey="id"
          size="small"
        />
      </Card>
    </div>
  );
};

export default SecurityLogs;
