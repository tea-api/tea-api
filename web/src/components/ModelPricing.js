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
  Button,
  Switch,
  InputNumber,
} from '@douyinfe/semi-ui';
import {
  IconMore,
  IconVerify,
  IconUploadError,
  IconHelpCircle,
  IconEdit,
  IconSave,
  IconRefresh,
} from '@douyinfe/semi-icons';
import { UserContext } from '../context/User/index.js';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';

const ModelPricing = () => {
  const { t } = useTranslation();
  const [filteredValue, setFilteredValue] = useState([]);
  const compositionRef = useRef({ isComposition: false });
  const [selectedRowKeys, setSelectedRowKeys] = useState([]);
  const [modalImageUrl, setModalImageUrl] = useState('');
  const [isModalOpenurl, setIsModalOpenurl] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState('default');
  const [editMode, setEditMode] = useState(false);
  const [editedPrices, setEditedPrices] = useState({});
  const [saving, setSaving] = useState(false);

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

  // 价格和倍率转换函数
  const calculateRatioFromPrice = (pricePerMillion) => {
    // 1倍率 = $0.002 / 1K tokens = $2 / 1M tokens
    return pricePerMillion / 2;
  };

  const calculatePriceFromRatio = (ratio) => {
    // 倍率 * 2 = $ / 1M tokens
    return ratio * 2;
  };

  // 处理价格编辑
  const handlePriceEdit = (modelName, field, value) => {
    const numValue = parseFloat(value) || 0;
    setEditedPrices(prev => ({
      ...prev,
      [modelName]: {
        ...prev[modelName],
        [field]: numValue
      }
    }));
  };

  // 获取编辑后的价格，如果没有编辑则返回原始计算值
  const getEditedPrice = (record, field) => {
    const editedModel = editedPrices[record.model_name];
    if (editedModel && editedModel[field] !== undefined) {
      return editedModel[field];
    }

    // 返回原始计算值
    if (record.quota_type === 0) {
      const basePrice = record.model_ratio * 2 * groupRatio[selectedGroup];
      if (field === 'inputPrice') {
        return parseFloat(basePrice.toFixed(6));
      } else if (field === 'outputPrice') {
        return parseFloat((basePrice * record.completion_ratio).toFixed(6));
      }
    }
    return 0;
  };

  // 初始化编辑价格（当切换到编辑模式时）
  const initializeEditPrices = () => {
    const initialPrices = {};
    models.forEach(model => {
      if (model.quota_type === 0) { // 只对按量计费的模型初始化
        const basePrice = model.model_ratio * 2 * groupRatio[selectedGroup];
        initialPrices[model.model_name] = {
          inputPrice: parseFloat(basePrice.toFixed(6)),
          outputPrice: parseFloat((basePrice * model.completion_ratio).toFixed(6))
        };
      }
    });
    setEditedPrices(initialPrices);
  };

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
        if (record.quota_type === 0) {
          // 按量计费模型
          if (editMode) {
            // 编辑模式
            const inputPrice = getEditedPrice(record, 'inputPrice');
            const outputPrice = getEditedPrice(record, 'outputPrice');

            return (
              <div style={{ minWidth: '200px' }}>
                <div style={{ marginBottom: '8px' }}>
                  <Text style={{ display: 'inline-block', width: '50px' }}>
                    {t('输入')}:
                  </Text>
                  <InputNumber
                    value={inputPrice}
                    onChange={(value) => handlePriceEdit(record.model_name, 'inputPrice', value)}
                    style={{ width: '120px' }}
                    suffix="$/1M"
                    precision={6}
                    step={0.001}
                  />
                </div>
                <div>
                  <Text style={{ display: 'inline-block', width: '50px' }}>
                    {t('输出')}:
                  </Text>
                  <InputNumber
                    value={outputPrice}
                    onChange={(value) => handlePriceEdit(record.model_name, 'outputPrice', value)}
                    style={{ width: '120px' }}
                    suffix="$/1M"
                    precision={6}
                    step={0.001}
                  />
                </div>
                {/* 显示计算后的倍率 */}
                <div style={{ marginTop: '4px', fontSize: '12px', color: '#666' }}>
                  <Text>
                    {t('模型倍率')}: {calculateRatioFromPrice(inputPrice).toFixed(3)}
                  </Text>
                  <br />
                  <Text>
                    {t('补全倍率')}: {inputPrice > 0 ? (outputPrice / inputPrice).toFixed(3) : '0'}
                  </Text>
                </div>
              </div>
            );
          } else {
            // 查看模式
            let inputRatioPrice =
              record.model_ratio * 2 * groupRatio[selectedGroup];
            let completionRatioPrice =
              record.model_ratio *
              record.completion_ratio *
              2 *
              groupRatio[selectedGroup];
            return (
              <>
                <Text>
                  {t('输入')} ${inputRatioPrice.toFixed(6)} / 1M tokens
                </Text>
                <br />
                <Text>
                  {t('输出')} ${completionRatioPrice.toFixed(6)} / 1M tokens
                </Text>
              </>
            );
          }
        } else {
          // 按次计费模型
          let price = parseFloat(text) * groupRatio[selectedGroup];
          return (
            <>
              <Text>{t('固定价格')}: ${price.toFixed(4)}</Text>
            </>
          );
        }
      },
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

    let url = '';
    url = `/api/pricing`;
    const res = await API.get(url);
    const { success, message, data, group_ratio, usable_group } = res.data;
    if (success) {
      setGroupRatio(group_ratio);
      setUsableGroup(usable_group);
      setSelectedGroup(userState.user ? userState.user.group : 'default');
      setModelsFormat(data, group_ratio);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const refresh = async () => {
    await loadPricing();
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess('已复制：' + text);
    } else {
      // setSearchKeyword(text);
      Modal.error({ title: '无法复制到剪贴板，请手动复制', content: text });
    }
  };

  // 保存价格修改
  const savePriceChanges = async () => {
    if (Object.keys(editedPrices).length === 0) {
      showInfo(t('没有需要保存的修改'));
      return;
    }

    setSaving(true);
    try {
      // 构建要更新的数据
      const updates = {
        modelRatios: {},
        completionRatios: {}
      };

      Object.entries(editedPrices).forEach(([modelName, prices]) => {
        if (prices.inputPrice !== undefined) {
          // 计算模型倍率 (去除分组倍率影响)
          const modelRatio = calculateRatioFromPrice(prices.inputPrice) / groupRatio[selectedGroup];
          updates.modelRatios[modelName] = modelRatio;
        }

        if (prices.outputPrice !== undefined && prices.inputPrice !== undefined && prices.inputPrice > 0) {
          // 计算补全倍率
          const completionRatio = prices.outputPrice / prices.inputPrice;
          updates.completionRatios[modelName] = completionRatio;
        }
      });

      // 发送更新请求
      const res = await API.post('/api/pricing/update', updates);
      const { success, message } = res.data;

      if (success) {
        showSuccess(t('价格更新成功'));
        setEditedPrices({});
        setEditMode(false);
        await refresh(); // 重新加载数据
      } else {
        showError(message || t('价格更新失败'));
      }
    } catch (error) {
      showError(t('价格更新失败: ') + error.message);
    } finally {
      setSaving(false);
    }
  };

  // 取消编辑
  const cancelEdit = () => {
    setEditedPrices({});
    setEditMode(false);
  };

  useEffect(() => {
    refresh().then();
  }, []);

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

          {/* 编辑模式控制 */}
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <Text>{t('编辑模式')}:</Text>
            <Switch
              checked={editMode}
              onChange={(checked) => {
                if (!checked) {
                  cancelEdit();
                } else {
                  setEditMode(true);
                  initializeEditPrices();
                }
              }}
              disabled={saving}
            />
          </div>

          {editMode && (
            <>
              <Button
                type='primary'
                icon={<IconSave />}
                onClick={savePriceChanges}
                loading={saving}
                disabled={Object.keys(editedPrices).length === 0}
              >
                {t('保存修改')}
              </Button>
              <Button
                icon={<IconRefresh />}
                onClick={cancelEdit}
                disabled={saving}
              >
                {t('取消')}
              </Button>
            </>
          )}
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
