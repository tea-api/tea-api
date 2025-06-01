import React, { useContext, useEffect, useState } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { StatusContext } from '../context/Status';
import { API } from '../helpers';

const SetupCheck = ({ children }) => {
  const [statusState] = useContext(StatusContext);
  const location = useLocation();
  const [setupChecked, setSetupChecked] = useState(false);
  const [setupRequired, setSetupRequired] = useState(false);

  useEffect(() => {
    const checkSetupStatus = async () => {
      try {
        // 直接调用 /api/setup 来获取准确的setup状态
        const res = await API.get('/api/setup');
        const { success, data } = res.data;

        if (success) {
          setSetupRequired(!data.status);
          setSetupChecked(true);

          // 如果需要setup且当前不在setup页面，则重定向
          if (!data.status && location.pathname !== '/setup') {
            window.location.href = '/setup';
          }
        } else {
          // 如果API调用失败，回退到使用status API的结果
          const setupFromStatus = statusState?.status?.setup;
          if (setupFromStatus === false && location.pathname !== '/setup') {
            window.location.href = '/setup';
          }
          setSetupChecked(true);
        }
      } catch (error) {
        console.error('Failed to check setup status:', error);
        // 如果setup API失败，回退到使用status API的结果
        const setupFromStatus = statusState?.status?.setup;
        if (setupFromStatus === false && location.pathname !== '/setup') {
          window.location.href = '/setup';
        }
        setSetupChecked(true);
      }
    };

    // 只有当statusState加载完成后才检查setup状态
    if (statusState?.status !== undefined) {
      checkSetupStatus();
    }
  }, [statusState?.status, location.pathname]);

  // 在setup状态检查完成前，显示children（避免闪烁）
  return children;
};

export default SetupCheck;