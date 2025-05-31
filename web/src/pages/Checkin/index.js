import React, { useState, useEffect } from 'react';
import { Button, Card, Spin, Typography, Space, Badge, Tag, Tooltip } from '@douyinfe/semi-ui';
import { IconTickCircle, IconCalendar, IconGift, IconCrown, IconHelpCircle } from '@douyinfe/semi-icons';
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
  const [specialRewards, setSpecialRewards] = useState([]);
  const [weeklyBonus, setWeeklyBonus] = useState(0);
  const [monthlyBonus, setMonthlyBonus] = useState(0);
  const [specialDays, setSpecialDays] = useState([]);

  useEffect(() => {
    const checkStatus = async () => {
      try {
        // 获取签到配置
        const configRes = await API.get('/api/checkin/config');
        if (configRes.data.success) {
          const { data } = configRes.data;
          // 检查签到功能是否启用
          if (!data.checkin_enabled) {
            showError(t('签到功能未启用，请联系管理员'));
            setLoading(false);
            return;
          }
          // 获取签到奖励配置
          if (data.checkin_config) {
            setBaseReward(data.checkin_config.base_reward || 10000);
            setContinuousReward(data.checkin_config.continuous_reward || 1000);
            setMaxDays(data.checkin_config.max_days || 7);
            setWeeklyBonus(data.checkin_config.weekly_bonus || 20000);
            setMonthlyBonus(data.checkin_config.monthly_bonus || 50000);
            
            // 设置特殊日期奖励
            if (data.checkin_config.special_rewards) {
              setSpecialRewards(data.checkin_config.special_rewards);
            }
            
            // 设置特殊日期
            if (data.checkin_config.special_days) {
              setSpecialDays(data.checkin_config.special_days);
            }
          }
        }

        // 获取签到状态
        const res = await API.get('/api/checkin/status');
        const { success, data } = res.data;
        if (success) {
          if (data.checked_today) {
            setDays(data.continuous);
            setReward(data.reward);
            setDone(true);
          } else {
            setDays(data.continuous || 0);
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
        
        let rewardMessage = `${t('今日获得')} ${data.reward.toLocaleString()} ${t('配额奖励')}`;
        if (data.special_reward) {
          rewardMessage += t('额外获得{{name}}奖励！', { name: data.special_reward.name });
        }
        showSuccess(t('签到成功') + '，' + rewardMessage);
        
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

  // 检查是否是特殊日期
  const isSpecialDay = (day) => {
    return specialDays.includes(day);
  };

  // 生成签到日历
  const renderCalendar = () => {
    const calendar = [];
    for (let i = 1; i <= maxDays; i++) {
      const isSpecial = isSpecialDay(i);
      calendar.push(
        <div key={i} className={`calendar-day ${i <= days ? 'checked' : ''} ${i === days && done ? 'today' : ''} ${isSpecial ? 'special-day' : ''}`}>
          <Badge dot={i <= days} type={isSpecial ? "danger" : "primary"}>
            <div className="day-circle">
              {isSpecial && <IconCrown className="special-icon" />}
              {i <= days ? <IconTickCircle /> : i}
            </div>
          </Badge>
          {isSpecial && (
            <div className="special-marker"></div>
          )}
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
            {days !== null && (
              <Text strong className="checkin-days">
                {t('已连续签到')} <span className="highlight">{days}</span> {t('天')}
              </Text>
            )}
            
            {done && reward > 0 && (
              <div className={`reward-container ${animate ? 'animate' : ''}`}>
                <IconGift size="large" />
                <Text className="reward-text">
                  {t('今日获得')} <span className="highlight">{formatQuota(reward)}</span> {t('配额奖励')}
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
              {t('签到奖励规则')}
            </Title>
            <Space vertical>
              <Text>1. {t('基础签到奖励')} <span className="highlight">{formatQuota(baseReward)}</span></Text>
              <Text>2. {t('连续签到额外奖励')} <span className="highlight">{formatQuota(continuousReward)}</span></Text>
              <Text>3. {t('最大连续奖励天数')} <span className="highlight">{maxDays}</span></Text>
              <Text>4. {t('中断签到重置连续计数')}</Text>
              
              <div className="special-rewards-section">
                <Text strong style={{ display: 'flex', alignItems: 'center' }}>
                  <IconCrown style={{ marginRight: '8px', color: '#FF9500' }} />
                  {t('特殊奖励')}
                </Text>
                <div className="special-rewards-list">
                  <div className="special-reward-item">
                    <Tag color="blue" size="large">{t('周奖励')}</Tag>
                    <Text>{t('连续签到7天可获得额外')} <span className="highlight">{formatQuota(weeklyBonus)}</span> {t('配额')}</Text>
                  </div>
                  <div className="special-reward-item">
                    <Tag color="purple" size="large">{t('月奖励')}</Tag>
                    <Text>{t('连续签到30天可获得额外')} <span className="highlight">{formatQuota(monthlyBonus)}</span> {t('配额')}</Text>
                  </div>
                  
                  {specialRewards.length > 0 && specialRewards.map((reward, index) => (
                    <div key={index} className="special-reward-item">
                      <Tag color="orange" size="large">
                        {reward.name}
                        <Tooltip content={reward.description}>
                          <IconHelpCircle style={{ marginLeft: '4px' }} />
                        </Tooltip>
                      </Tag>
                      <Text>{reward.description}</Text>
                    </div>
                  ))}
                </div>
              </div>
            </Space>
          </Card>
        </Space>
      )}
    </Card>
  );
};

export default Checkin;
