import React, { useContext, useEffect, useRef, useMemo, useState } from 'react';
import { API, copy, showError, showInfo, showSuccess } from '../helpers';
import { useTranslation } from 'react-i18next';

import {
  Banner,
  Input,
  Layout,
  Modal,
  Space,
  Table,
  Tag,
  Tooltip,
  Popover,
  ImagePreview,
  RadioGroup,
  Radio,
  Button,
} from '@douyinfe/semi-ui';
import {
  IconMore,
  IconVerify,
  IconUploadError,
  IconHelpCircle,
  IconEdit,
} from '@douyinfe/semi-icons';
import { UserContext } from '../context/User/index.js';
import { isAdmin } from '../helpers/utils.js';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';

const ModelPricing = () => {
  const { t } = useTranslation();
  const [filteredValue, setFilteredValue] = useState([]);
  const compositionRef = useRef({ isComposition: false });
  const [selectedRowKeys, setSelectedRowKeys] = useState([]);
  const [modalImageUrl, setModalImageUrl] = useState('');
  const [isModalOpenurl, setIsModalOpenurl] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState('default');
  const [editVisible, setEditVisible] = useState(false);
  const [editModel, setEditModel] = useState(null);
  const [inputPrice, setInputPrice] = useState('');
  const [outputPrice, setOutputPrice] = useState('');
  const [priceUnit, setPriceUnit] = useState('1m');

  const rowSelection = useMemo(
    () => ({
      onChange: (selectedRowKeys, selectedRows) => {
        setSelectedRowKeys(selectedRowKeys);
      },
    }),
    [],
  );

  const handleChange = (value) => {
    if (compositionRef.current.isComposition) {
      return;
    }
    const newFilteredValue = value ? [value] : [];
    setFilteredValue(newFilteredValue);
  };
  const handleCompositionStart = () => {
    compositionRef.current.isComposition = true;
  };

  const handleCompositionEnd = (event) => {
    compositionRef.current.isComposition = false;
    const value = event.target.value;
    const newFilteredValue = value ? [value] : [];
    setFilteredValue(newFilteredValue);
  };

  const openEdit = (record) => {
    // 先设置单位为默认值
    setPriceUnit('1m');
    
    // 使用与后端一致的基准价格计算方式
    const basePerM = 2.0;
    
    // 根据模型类型设置初始价格
    if (record.quota_type === 0) {
      // 按量计费模型
      // 显示每1M tokens的价格
      setInputPrice((record.model_ratio * basePerM).toFixed(3));
      setOutputPrice((record.model_ratio * record.completion_ratio * basePerM).toFixed(3));
    } else {
      // 按次计费模型
      setInputPrice(record.model_price.toFixed(3));
      setOutputPrice('0');
    }
    
    setEditModel(record);
    setEditVisible(true);
  };

  const savePrice = async () => {
    if (!editModel) return;
    
    try {
      // 更新模型价格
      console.log(`准备更新模型 ${editModel.model_name} 的价格，输入价格: ${inputPrice}，输出价格: ${outputPrice}`);
      
      const res = await API.put('/api/pricing', {
        model_name: editModel.model_name,
        input_price: parseFloat(inputPrice),
        output_price: parseFloat(outputPrice),
        unit: priceUnit,
      });
      
      console.log('后端返回结果:', res.data);
      
      if (res.data.success) {
        // 使用后端返回的补全倍率
        if (editModel.quota_type === 0 && res.data.completion_ratio !== undefined) {
          const completionRatio = res.data.completion_ratio;
          console.log(`后端返回的补全倍率: ${completionRatio}`);
          showSuccess(t('保存成功，补全倍率已设为: {{ratio}}', { 
            ratio: completionRatio.toFixed(3) 
          }));
        } else {
          console.log('未获取到后端返回的补全倍率或非按量计费模型');
          showSuccess(t('保存成功'));
        }
        
        setEditVisible(false);
        // 刷新数据，但不重新加载页面
        console.log('开始刷新数据...');
        await refresh();
        console.log('数据刷新完成');
      } else {
        showError(res.data.message || t('保存失败'));
      }
    } catch (error) {
      console.error('保存价格时出错:', error);
      showError(t('保存价格失败，请稍后重试'));
    }
  };

  function renderQuotaType(type) {
    // Ensure all cases are string literals by adding quotes.
    switch (type) {
      case 1:
        return (
          <Tag color='teal' size='large'>
            {t('按次计费')}
          </Tag>
        );
      case 0:
        return (
          <Tag color='violet' size='large'>
            {t('按量计费')}
          </Tag>
        );
      default:
        return t('未知');
    }
  }

  function renderAvailable(available) {
    return available ? (
      <Popover
        content={
          <div style={{ padding: 8 }}>{t('您的分组可以使用该模型')}</div>
        }
        position='top'
        key={available}
        style={{
          backgroundColor: 'rgba(var(--semi-blue-4),1)',
          borderColor: 'rgba(var(--semi-blue-4),1)',
          color: 'var(--semi-color-white)',
          borderWidth: 1,
          borderStyle: 'solid',
        }}
      >
        <IconVerify style={{ color: 'green' }} size='large' />
      </Popover>
    ) : null;
  }

  const columns = [
    {
      title: t('可用性'),
      dataIndex: 'available',
      render: (text, record, index) => {
        // if record.enable_groups contains selectedGroup, then available is true
        return renderAvailable(record.enable_groups.includes(selectedGroup));
      },
      sorter: (a, b) => {
        const aAvailable = a.enable_groups.includes(selectedGroup);
        const bAvailable = b.enable_groups.includes(selectedGroup);
        return Number(aAvailable) - Number(bAvailable);
      },
      defaultSortOrder: 'descend',
    },
    {
      title: t('模型名称'),
      dataIndex: 'model_name',
      render: (text, record, index) => {
        return (
          <>
            <Tag
              color='green'
              size='large'
              onClick={() => {
                copyText(text);
              }}
            >
              {text}
            </Tag>
          </>
        );
      },
      onFilter: (value, record) =>
        record.model_name.toLowerCase().includes(value.toLowerCase()),
      filteredValue,
    },
    {
      title: t('计费类型'),
      dataIndex: 'quota_type',
      render: (text, record, index) => {
        return renderQuotaType(parseInt(text));
      },
      sorter: (a, b) => a.quota_type - b.quota_type,
    },
    {
      title: t('可用分组'),
      dataIndex: 'enable_groups',
      render: (text, record, index) => {
        // enable_groups is a string array
        return (
          <Space>
            {text.map((group) => {
              if (usableGroup[group]) {
                if (group === selectedGroup) {
                  return (
                    <Tag color='blue' size='large' prefixIcon={<IconVerify />}>
                      {group}
                    </Tag>
                  );
                } else {
                  return (
                    <Tag
                      color='blue'
                      size='large'
                      onClick={() => {
                        setSelectedGroup(group);
                        showInfo(
                          t('当前查看的分组为：{{group}}，倍率为：{{ratio}}', {
                            group: group,
                            ratio: groupRatio[group],
                          }),
                        );
                      }}
                    >
                      {group}
                    </Tag>
                  );
                }
              }
            })}
          </Space>
        );
      },
    },
    {
      title: () => (
        <span style={{ display: 'flex', alignItems: 'center' }}>
          {t('倍率')}
          <Popover
            content={
              <div style={{ padding: 8 }}>
                {t('倍率是为了方便换算不同价格的模型')}
                <br />
                {t('点击查看倍率说明')}
              </div>
            }
            position='top'
            style={{
              backgroundColor: 'rgba(var(--semi-blue-4),1)',
              borderColor: 'rgba(var(--semi-blue-4),1)',
              color: 'var(--semi-color-white)',
              borderWidth: 1,
              borderStyle: 'solid',
            }}
          >
            <IconHelpCircle
              onClick={() => {
                setModalImageUrl('/ratio.png');
                setIsModalOpenurl(true);
              }}
            />
          </Popover>
        </span>
      ),
      dataIndex: 'model_ratio',
      render: (text, record, index) => {
        let content = text;
        let completionRatio = parseFloat(record.completion_ratio.toFixed(3));
        content = (
          <>
            <Text>
              {t('模型倍率')}：{record.quota_type === 0 ? text : t('无')}
            </Text>
            <br />
            <Text>
              {t('补全倍率')}：
              {record.quota_type === 0 ? completionRatio : t('无')}
            </Text>
            <br />
            <Text>
              {t('分组倍率')}：{groupRatio[selectedGroup]}
            </Text>
          </>
        );
        return <div>{content}</div>;
      },
    },
    {
      title: t('模型价格'),
      dataIndex: 'model_price',
      render: (text, record, index) => {
        let content = text;
        if (record.quota_type === 0) {
          // 按量计费模型
          // 使用与后端一致的基准价格计算方式 - 1倍率=0.002刀
          const basePerM = 2.0;
          const base = basePerM / 1000; // 转换为每千 tokens 的价格
          
          let inputRatioPrice = record.model_ratio * base * groupRatio[selectedGroup];
          let completionRatioPrice = record.model_ratio * record.completion_ratio * base * groupRatio[selectedGroup];
          
          // 显示为每1M tokens的价格
          let inputRatioPricePerM = inputRatioPrice * 1000;
          let completionRatioPricePerM = completionRatioPrice * 1000;
          
          content = (
            <>
              <Text>
                {t('提示')} ${inputRatioPricePerM.toFixed(3)} / 1M tokens
              </Text>
              <br />
              <Text>
                {t('补全')} ${completionRatioPricePerM.toFixed(3)} / 1M tokens
              </Text>
            </>
          );
        } else {
          // 按次计费模型
          let price = parseFloat(text) * groupRatio[selectedGroup];
          content = (
            <>
              {t('模型价格')}：${price.toFixed(3)}
            </>
          );
        }
        return <div>{content}</div>;
      },
    },
    {
      title: t('编辑价格'),
      dataIndex: 'action',
      render: (text, record) => (
        isAdmin() && (
          <Space>
            <Button icon={<IconEdit />} onClick={() => openEdit(record)} />
          </Space>
        )
      ),
    },
  ];

  const [models, setModels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [userState, userDispatch] = useContext(UserContext);
  const [groupRatio, setGroupRatio] = useState({});
  const [usableGroup, setUsableGroup] = useState({});

  const setModelsFormat = (models, groupRatio) => {
    for (let i = 0; i < models.length; i++) {
      models[i].key = models[i].model_name;
      models[i].group_ratio = groupRatio[models[i].model_name];
    }
    // sort by quota_type
    models.sort((a, b) => {
      return a.quota_type - b.quota_type;
    });

    // sort by model_name, start with gpt is max, other use localeCompare
    models.sort((a, b) => {
      if (a.model_name.startsWith('gpt') && !b.model_name.startsWith('gpt')) {
        return -1;
      } else if (
        !a.model_name.startsWith('gpt') &&
        b.model_name.startsWith('gpt')
      ) {
        return 1;
      } else {
        return a.model_name.localeCompare(b.model_name);
      }
    });

    setModels(models);
  };

  const loadPricing = async () => {
    setLoading(true);
    console.log('开始加载价格数据...');

    // 添加时间戳参数避免缓存
    const timestamp = new Date().getTime();
    let url = `/api/pricing?t=${timestamp}`;
    
    try {
      // 使用no-cache确保不使用缓存
      const res = await API.get(url, {
        headers: {
          'Cache-Control': 'no-cache',
          'Pragma': 'no-cache'
        }
      });
      const { success, message, data, group_ratio, usable_group } = res.data;
      if (success) {
        console.log('价格数据加载成功，模型数量:', data.length);
        setGroupRatio(group_ratio);
        setUsableGroup(usable_group);
        setSelectedGroup(userState.user ? userState.user.group : 'default');
        
        // 打印一些关键模型的补全倍率，用于调试
        if (data.length > 0) {
          const sampleModels = data.slice(0, Math.min(3, data.length));
          console.log('示例模型补全倍率:', sampleModels.map(m => ({ 
            name: m.model_name, 
            ratio: m.model_ratio,
            completionRatio: m.completion_ratio 
          })));
          
          // 如果存在刚刚编辑的模型，打印其信息
          if (editModel) {
            const editedModel = data.find(m => m.model_name === editModel.model_name);
            if (editedModel) {
              console.log('刚编辑的模型最新信息:', {
                name: editedModel.model_name,
                ratio: editedModel.model_ratio,
                completionRatio: editedModel.completion_ratio
              });
            }
          }
        }
        
        setModelsFormat(data, group_ratio);
      } else {
        showError(message);
        console.error('加载价格数据失败:', message);
      }
    } catch (error) {
      console.error('加载价格数据出错:', error);
      showError(t('加载价格数据失败'));
    }
    
    setLoading(false);
  };

  const refresh = async () => {
    console.log('执行刷新操作...');
    // 强制清除浏览器缓存
    await loadPricing();
    console.log('刷新完成');
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess('已复制：' + text);
    } else {
      // setSearchKeyword(text);
      Modal.error({ title: '无法复制到剪贴板，请手动复制', content: text });
    }
  };

  useEffect(() => {
    refresh().then();
  }, []);

  // 处理输入价格变化
  const handleInputPriceChange = (value) => {
    setInputPrice(value);
    // 实时更新Banner显示的补全倍率
    updateCompletionRatioBanner(value, outputPrice);
  };
  
  // 处理输出价格变化
  const handleOutputPriceChange = (value) => {
    setOutputPrice(value);
    // 实时更新Banner显示的补全倍率
    updateCompletionRatioBanner(inputPrice, value);
  };
  
  // 更新补全倍率显示
  const updateCompletionRatioBanner = (input, output) => {
    const inputVal = parseFloat(input);
    const outputVal = parseFloat(output);
    
    if (!isNaN(inputVal) && !isNaN(outputVal) && inputVal > 0) {
      const ratio = outputVal / inputVal;
      console.log('实时计算的补全倍率:', ratio.toFixed(3));
    }
  };

  // 处理单位切换
  const handleUnitChange = (value) => {
    if (!editModel) return;
    
    const oldUnit = priceUnit;
    const newUnit = value;
    
    // 如果单位不同，需要转换价格
    if (oldUnit !== newUnit) {
      const currentInputPrice = parseFloat(inputPrice);
      const currentOutputPrice = parseFloat(outputPrice);
      
      let newInputPrice, newOutputPrice;
      
      if (oldUnit === '1k' && newUnit === '1m') {
        // 从1k转换到1M，价格乘以1000
        newInputPrice = (currentInputPrice * 1000).toFixed(3);
        newOutputPrice = (currentOutputPrice * 1000).toFixed(3);
      } else if (oldUnit === '1m' && newUnit === '1k') {
        // 从1M转换到1k，价格除以1000
        newInputPrice = (currentInputPrice / 1000).toFixed(3);
        newOutputPrice = (currentOutputPrice / 1000).toFixed(3);
      }
      
      setInputPrice(newInputPrice);
      setOutputPrice(newOutputPrice);
      
      // 保留计算逻辑，但删除showInfo
    }
    
    setPriceUnit(value);
  };

  return (
    <>
      <Layout>
        {userState.user ? (
          <Banner
            type='success'
            fullMode={false}
            closeIcon='null'
            description={t('您的默认分组为：{{group}}，分组倍率为：{{ratio}}', {
              group: userState.user.group,
              ratio: groupRatio[userState.user.group],
            })}
          />
        ) : (
          <Banner
            type='warning'
            fullMode={false}
            closeIcon='null'
            description={t('您还未登陆，显示的价格为默认分组倍率: {{ratio}}', {
              ratio: groupRatio['default'],
            })}
          />
        )}
        <br />
        <Banner
          type='info'
          fullMode={false}
          description={
            <div>
              {t(
                '按量计费费用 = 分组倍率 × 模型倍率 × （提示token数 + 补全token数 × 补全倍率）/ 500000 （单位：美元）',
              )}
            </div>
          }
          closeIcon='null'
        />
        <br />
        <Space style={{ marginBottom: 16 }}>
          <Input
            placeholder={t('模糊搜索模型名称')}
            style={{ width: 200 }}
            onCompositionStart={handleCompositionStart}
            onCompositionEnd={handleCompositionEnd}
            onChange={handleChange}
            showClear
          />
          <Button
            theme='light'
            type='tertiary'
            style={{ width: 150 }}
            onClick={() => {
              copyText(selectedRowKeys);
            }}
            disabled={selectedRowKeys == ''}
          >
            {t('复制选中模型')}
          </Button>
        </Space>
        <Table
          style={{ marginTop: 5 }}
          columns={columns}
          dataSource={models}
          loading={loading}
          pagination={{
            formatPageText: (page) =>
              t('第 {{start}} - {{end}} 条，共 {{total}} 条', {
                start: page.currentStart,
                end: page.currentEnd,
                total: models.length,
              }),
            pageSize: models.length,
            showSizeChanger: false,
          }}
          rowSelection={rowSelection}
        />
        <Modal
          visible={editVisible}
          title={t('编辑价格')}
          onOk={savePrice}
          onCancel={() => setEditVisible(false)}
        >
          <RadioGroup
            type='button'
            value={priceUnit}
            onChange={(e) => handleUnitChange(e.target.value)}
            style={{ marginBottom: 12 }}
          >
            <Radio value='1k'>/1k tokens</Radio>
            <Radio value='1m'>/1M tokens</Radio>
          </RadioGroup>
          <Input
            value={inputPrice}
            onChange={handleInputPriceChange}
            suffix={`$/1${priceUnit === '1k' ? 'k' : 'M'} tokens`}
            label={t('实际输入价格')}
          />
          <Input
            value={outputPrice}
            onChange={handleOutputPriceChange}
            suffix={`$/1${priceUnit === '1k' ? 'k' : 'M'} tokens`}
            label={t('实际输出价格')}
            style={{ marginTop: 12 }}
          />
        </Modal>
        <ImagePreview
          src={modalImageUrl}
          visible={isModalOpenurl}
          onVisibleChange={(visible) => setIsModalOpenurl(visible)}
        />
      </Layout>
    </>
  );
};

export default ModelPricing;
