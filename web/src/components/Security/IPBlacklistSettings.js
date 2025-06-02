import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Form, 
  Row, 
  Col, 
  Button, 
  Switch, 
  InputNumber, 
  Typography, 
  Space, 
  Divider,
  Table,
  Input,
  Modal,
  Tag,
  Popconfirm
} from '@douyinfe/semi-ui';
import { IconLock, IconSave, IconPlus, IconDelete } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';

const { Title, Text } = Typography;

const IPBlacklistSettings = ({ config, refresh }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [addModalVisible, setAddModalVisible] = useState(false);
  const [whitelistModalVisible, setWhitelistModalVisible] = useState(false);
  const [blacklistData, setBlacklistData] = useState([]);
  const [formData, setFormData] = useState({
    enabled: true,
    temp_block_duration_hours: 1,
    perm_block_duration_hours: 24,
    max_violations: 5,
    cleanup_interval_minutes: 10,
    auto_blacklist_enabled: true
  });
  const [newIP, setNewIP] = useState('');
  const [newReason, setNewReason] = useState('');
  const [isTemporary, setIsTemporary] = useState(true);

  useEffect(() => {
    if (config?.ip_blacklist) {
      setFormData({
        enabled: config.ip_blacklist.enabled || false,
        temp_block_duration_hours: config.ip_blacklist.temp_block_duration_hours || 1,
        perm_block_duration_hours: config.ip_blacklist.perm_block_duration_hours || 24,
        max_violations: config.ip_blacklist.max_violations || 5,
        cleanup_interval_minutes: config.ip_blacklist.cleanup_interval_minutes || 10,
        auto_blacklist_enabled: config.ip_blacklist.auto_blacklist_enabled !== false
      });
    }
    loadBlacklistData();
  }, [config]);

  const loadBlacklistData = async () => {
    try {
      const response = await API.get('/api/security/blacklist');
      if (response.data.success) {
        // 这里应该返回实际的黑名单数据，目前返回模拟数据
        setBlacklistData([
          {
            ip: '192.168.1.100',
            reason: 'Token浪费攻击',
            blocked_at: '2024-01-01T12:00:00Z',
            expires_at: '2024-01-01T13:00:00Z',
            violations: 3,
            is_temporary: true
          }
        ]);
      }
    } catch (error) {
      console.error('Failed to load blacklist data:', error);
    }
  };

  const handleSave = async () => {
    setLoading(true);
    try {
      const newConfig = {
        ...config,
        ip_blacklist: {
          enabled: formData.enabled,
          temp_block_duration_hours: formData.temp_block_duration_hours,
          perm_block_duration_hours: formData.perm_block_duration_hours,
          max_violations: formData.max_violations,
          cleanup_interval_minutes: formData.cleanup_interval_minutes,
          auto_blacklist_enabled: formData.auto_blacklist_enabled
        }
      };

      const response = await API.put('/api/security/config', newConfig);
      
      if (response.data.success) {
        showSuccess(t('IP黑名单设置保存成功'));
        refresh();
      } else {
        showError(response.data.message || t('保存失败'));
      }
    } catch (error) {
      showError(t('保存IP黑名单设置失败'));
    } finally {
      setLoading(false);
    }
  };

  const handleAddToBlacklist = async () => {
    if (!newIP || !newReason) {
      showError(t('请填写IP地址和封禁原因'));
      return;
    }

    try {
      const response = await API.post('/api/security/blacklist', {
        ip: newIP,
        reason: newReason,
        temporary: isTemporary
      });

      if (response.data.success) {
        showSuccess(t('IP已添加到黑名单'));
        setAddModalVisible(false);
        setNewIP('');
        setNewReason('');
        loadBlacklistData();
      } else {
        showError(response.data.message || t('添加失败'));
      }
    } catch (error) {
      showError(t('添加IP到黑名单失败'));
    }
  };

  const handleRemoveFromBlacklist = async (ip) => {
    try {
      const response = await API.delete(`/api/security/blacklist/${ip}`);
      
      if (response.data.success) {
        showSuccess(t('IP已从黑名单移除'));
        loadBlacklistData();
      } else {
        showError(response.data.message || t('移除失败'));
      }
    } catch (error) {
      showError(t('移除IP失败'));
    }
  };

  const handleFieldChange = (field, value) => {
    setFormData(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const columns = [
    {
      title: t('IP地址'),
      dataIndex: 'ip',
      key: 'ip',
      width: 150
    },
    {
      title: t('封禁原因'),
      dataIndex: 'reason',
      key: 'reason',
      width: 200
    },
    {
      title: t('类型'),
      dataIndex: 'is_temporary',
      key: 'type',
      width: 100,
      render: (isTemp) => (
        <Tag color={isTemp ? 'orange' : 'red'}>
          {isTemp ? t('临时') : t('永久')}
        </Tag>
      )
    },
    {
      title: t('违规次数'),
      dataIndex: 'violations',
      key: 'violations',
      width: 100
    },
    {
      title: t('封禁时间'),
      dataIndex: 'blocked_at',
      key: 'blocked_at',
      width: 180,
      render: (time) => new Date(time).toLocaleString()
    },
    {
      title: t('操作'),
      key: 'action',
      width: 120,
      render: (_, record) => (
        <Popconfirm
          title={t('确定要移除此IP吗？')}
          onConfirm={() => handleRemoveFromBlacklist(record.ip)}
        >
          <Button
            type="danger"
            icon={<IconDelete />}
            size="small"
          >
            {t('移除')}
          </Button>
        </Popconfirm>
      )
    }
  ];

  return (
    <div>
      <Card>
        <div style={{ marginBottom: '24px' }}>
          <Space>
            <IconLock size="large" style={{ color: '#ff4d4f' }} />
            <Title heading={4} style={{ margin: 0 }}>
              {t('IP黑名单管理')}
            </Title>
          </Space>
          <Text type="secondary" style={{ display: 'block', marginTop: '8px' }}>
            {t('管理被封禁的IP地址，防止恶意访问')}
          </Text>
        </div>

        <Form layout="vertical">
          {/* 基础设置 */}
          <Card title={t('基础设置')} style={{ marginBottom: '24px' }}>
            <Row gutter={24}>
              <Col span={12}>
                <Form.Item label={t('启用IP黑名单')}>
                  <Switch
                    checked={formData.enabled}
                    onChange={(checked) => handleFieldChange('enabled', checked)}
                    checkedText={t('开启')}
                    uncheckedText={t('关闭')}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('总开关，关闭后IP黑名单功能将停用')}
                  </Text>
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={t('自动加入黑名单')}>
                  <Switch
                    checked={formData.auto_blacklist_enabled}
                    onChange={(checked) => handleFieldChange('auto_blacklist_enabled', checked)}
                    checkedText={t('开启')}
                    uncheckedText={t('关闭')}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('检测到恶意行为时自动加入黑名单')}
                  </Text>
                </Form.Item>
              </Col>
            </Row>
          </Card>

          {/* 封禁时长设置 */}
          <Card title={t('封禁时长设置')} style={{ marginBottom: '24px' }}>
            <Row gutter={24}>
              <Col span={8}>
                <Form.Item label={t('临时封禁时长')}>
                  <InputNumber
                    value={formData.temp_block_duration_hours}
                    onChange={(value) => handleFieldChange('temp_block_duration_hours', value)}
                    min={1}
                    max={168}
                    step={1}
                    suffix={t('小时')}
                    style={{ width: '100%' }}
                  />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label={t('永久封禁时长')}>
                  <InputNumber
                    value={formData.perm_block_duration_hours}
                    onChange={(value) => handleFieldChange('perm_block_duration_hours', value)}
                    min={1}
                    max={8760}
                    step={1}
                    suffix={t('小时')}
                    style={{ width: '100%' }}
                  />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item label={t('最大违规次数')}>
                  <InputNumber
                    value={formData.max_violations}
                    onChange={(value) => handleFieldChange('max_violations', value)}
                    min={1}
                    max={20}
                    step={1}
                    suffix={t('次')}
                    style={{ width: '100%' }}
                  />
                  <Text type="secondary" style={{ display: 'block', marginTop: '4px' }}>
                    {t('超过此次数将转为永久封禁')}
                  </Text>
                </Form.Item>
              </Col>
            </Row>
          </Card>

          <Divider />

          <div style={{ textAlign: 'right', marginBottom: '24px' }}>
            <Button
              type="primary"
              icon={<IconSave />}
              loading={loading}
              onClick={handleSave}
              size="large"
            >
              {t('保存设置')}
            </Button>
          </div>
        </Form>
      </Card>

      {/* 黑名单管理 */}
      <Card>
        <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Title heading={5} style={{ margin: 0 }}>
            {t('黑名单列表')}
          </Title>
          <Button
            type="primary"
            icon={<IconPlus />}
            onClick={() => setAddModalVisible(true)}
          >
            {t('添加IP')}
          </Button>
        </div>

        <Table
          columns={columns}
          dataSource={blacklistData}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true
          }}
          rowKey="ip"
        />
      </Card>

      {/* 添加IP模态框 */}
      <Modal
        title={t('添加IP到黑名单')}
        visible={addModalVisible}
        onCancel={() => setAddModalVisible(false)}
        onOk={handleAddToBlacklist}
        okText={t('添加')}
        cancelText={t('取消')}
      >
        <Form layout="vertical">
          <Form.Item label={t('IP地址')} required>
            <Input
              value={newIP}
              onChange={setNewIP}
              placeholder={t('请输入IP地址')}
            />
          </Form.Item>
          <Form.Item label={t('封禁原因')} required>
            <Input
              value={newReason}
              onChange={setNewReason}
              placeholder={t('请输入封禁原因')}
            />
          </Form.Item>
          <Form.Item label={t('封禁类型')}>
            <Switch
              checked={isTemporary}
              onChange={setIsTemporary}
              checkedText={t('临时封禁')}
              uncheckedText={t('永久封禁')}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default IPBlacklistSettings;
