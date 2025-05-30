import React, { useState, useEffect } from 'react';
import { Button, Card, Spin, Typography, Space, Badge } from '@douyinfe/semi-ui';
import { IconTickCircle, IconCalendar, IconGift } from '@douyinfe/semi-icons';
import { API, showError, showSuccess } from '../../helpers';
import { useTranslation } from 'react-i18next';
import './style.css';

const { Text, Title } = Typography;

const Checkin = () => {
  const { t } = useTranslation();
  const [days, setDays] = useState(null);
  const [reward, setReward] = useState(0);
  const [done, setDone] = useState(false);
  const [loading, setLoading] = useState(true);
  const [baseReward, setBaseReward] = useState(10000);
  const [continuousReward, setContinuousReward] = useState(1000);
  const [maxDays, setMaxDays] = useState(7);
  const [animate, setAnimate] = useState(false);

  useEffect(() => {
    const checkStatus = async () => {
      try {
        // 获取签到配置
        const configRes = await API.get('/api/checkin/config');
        if (configRes.data.success) {
          const { data } = configRes.data;
          // 检查签到功能是否启用
          if (!data.checkin_enabled) {
            showError('签到功能未启用，请联系管理员');
            setLoading(false);
            return;
          }
          // 获取签到奖励配置
          if (data.checkin_config) {
            setBaseReward(data.checkin_config.base_reward || 10000);
            setContinuousReward(data.checkin_config.continuous_reward || 1000);
            setMaxDays(data.checkin_config.max_days || 7);
          }
        }

        // 获取签到状态
        const res = await API.get('/api/checkin/status');
        const { success, data } = res.data;
        if (success) {
          if (data.checked_today) {
            setDays(data.continuous);
            setDone(true);
          }
        }
      } catch (err) {
        console.error('获取签到状态失败', err);
      } finally {
        setLoading(false);
      }
    };
    
    checkStatus();
  }, []);

  const handleCheckin = async () => {
    try {
      const res = await API.post('/api/checkin');
      const { success, message, data } = res.data;
      if (success) {
        setDays(data.continuous);
        setReward(data.reward);
        setDone(true);
        setAnimate(true);
        showSuccess(t('签到成功') + `，获得 ${data.reward.toLocaleString()} 配额`);
        
        // 3秒后停止动画
        setTimeout(() => {
          setAnimate(false);
        }, 3000);
      } else {
        showError(message || t('签到失败'));
      }
    } catch (err) {
      showError(t('签到失败'));
    }
  };

  const formatQuota = (quota) => {
    return quota ? quota.toLocaleString() : '0';
  };

  // 生成签到日历
  const renderCalendar = () => {
    const calendar = [];
    for (let i = 1; i <= maxDays; i++) {
      calendar.push(
        <div key={i} className={`calendar-day ${i <= days ? 'checked' : ''} ${i === days && done ? 'today' : ''}`}>
          <Badge dot={i <= days} type="primary">
            <div className="day-circle">
              {i <= days ? <IconTickCircle /> : i}
            </div>
          </Badge>
        </div>
      );
    }
    return calendar;
  };

  return (
    <Card 
      className="checkin-card"
      title={
        <Title heading={4}>
          <IconCalendar style={{ marginRight: '8px' }} />
          {t('每日签到')}
        </Title>
      }
    >
      {loading ? (
        <div className="loading-container">
          <Spin size="large" />
        </div>
      ) : (
        <Space vertical align="center" spacing="loose" className="checkin-container">
          <div className="calendar-container">
            {renderCalendar()}
          </div>
          
          <div className="checkin-status">
            {days ? (
              <Text strong className="checkin-days">
                已连续签到 <span className="highlight">{days}</span> 天
              </Text>
            ) : (
              <Text strong>快来签到吧！</Text>
            )}
            
            {done && reward > 0 && (
              <div className={`reward-container ${animate ? 'animate' : ''}`}>
                <IconGift size="large" />
                <Text className="reward-text">
                  今日获得 <span className="highlight">{formatQuota(reward)}</span> 配额奖励
                </Text>
              </div>
            )}
          </div>
          
          <Button 
            theme='solid' 
            onClick={handleCheckin} 
            disabled={done} 
            size="large"
            className={`checkin-button ${done ? 'checked' : ''}`}
            icon={done ? <IconTickCircle /> : <IconCalendar />}
          >
            {done ? t('已签到') : t('签到')}
          </Button>
          
          <Card className="reward-rules">
            <Title heading={6} style={{ marginBottom: '12px' }}>
              <IconGift style={{ marginRight: '8px' }} />
              签到奖励规则
            </Title>
            <Space vertical>
              <Text>1. 基础奖励：每次签到获得 <span className="highlight">{formatQuota(baseReward)}</span> 配额</Text>
              <Text>2. 连续签到：连续签到每天额外奖励 <span className="highlight">{formatQuota(continuousReward)}</span> 配额</Text>
              <Text>3. 最大累计：连续签到奖励最多累计 <span className="highlight">{maxDays}</span> 天</Text>
              <Text>4. 中断计算：如果中断签到，连续天数将重新计算</Text>
            </Space>
          </Card>
        </Space>
      )}
    </Card>
  );
};

export default Checkin;
